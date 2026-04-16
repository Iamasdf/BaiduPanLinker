package handler

import (
	"io"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/qjfoidnh/BaiduPCS-Go/baidupcs"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
	"github.com/qjfoidnh/BaiduPCS-Go/pcsutil/converter"
)

func ResponseJSON(c *gin.Context, code int, message string, data interface{}) {
	c.JSON(200, gin.H{
		"code":    code,
		"message": message,
		"data":    data,
	})
}

func ResponseSuccess(c *gin.Context, data interface{}) {
	ResponseJSON(c, 200, "success", data)
}

func ResponseError(c *gin.Context, code int, message string) {
	ResponseJSON(c, code, message, nil)
}

type FileInfo struct {
	FsID     int64  `json:"fs_id"`
	Name     string `json:"name"`
	Path     string `json:"path"`
	Size     int64  `json:"size"`
	SizeStr  string `json:"size_str"`
	IsDir    bool   `json:"is_dir"`
	Mtime    int64  `json:"mtime"`
	MtimeStr string `json:"mtime_str"`
	MD5      string `json:"md5,omitempty"`
}

type DownloadLinkInfo struct {
	URL     string `json:"url"`
	Expired int64  `json:"expired"`
}

type FilesResponse struct {
	Path    string     `json:"path"`
	Files   []FileInfo `json:"files"`
	Total   int        `json:"total"`
	FileNum int        `json:"file_num"`
	DirNum  int        `json:"dir_num"`
}

func GetFiles(c *gin.Context) {
	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser == nil || activeUser.UID == 0 {
		ResponseError(c, 401, "未登录百度账号")
		return
	}

	path := c.DefaultQuery("path", "/")
	if path == "" {
		path = "/"
	}

	pcs := pcsconfig.Config.ActiveUserBaiduPCS()
	if pcs == nil {
		ResponseError(c, 500, "获取网盘实例失败")
		return
	}

	orderOptions := &baidupcs.OrderOptions{
		By:    baidupcs.OrderByName,
		Order: baidupcs.OrderAsc,
	}

	files, err := pcs.FilesDirectoriesList(path, orderOptions)
	if err != nil {
		ResponseError(c, 500, "获取文件列表失败: "+err.Error())
		return
	}

	fileInfos := make([]FileInfo, 0, len(files))
	fileNum, dirNum := 0, 0

	for _, f := range files {
		if f == nil {
			continue
		}
		info := FileInfo{
			FsID:     f.FsID,
			Name:     f.Filename,
			Path:     f.Path,
			Size:     f.Size,
			SizeStr:  converter.ConvertFileSize(f.Size, 2),
			IsDir:    f.Isdir,
			Mtime:    f.Mtime,
			MtimeStr: formatTime(f.Mtime),
		}
		if !f.Isdir && len(f.MD5) > 0 {
			info.MD5 = f.MD5
		}
		fileInfos = append(fileInfos, info)

		if f.Isdir {
			dirNum++
		} else {
			fileNum++
		}
	}

	response := FilesResponse{
		Path:    path,
		Files:   fileInfos,
		Total:   len(fileInfos),
		FileNum: fileNum,
		DirNum:  dirNum,
	}

	ResponseSuccess(c, response)
}

func GetDownloadLink(c *gin.Context) {
	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser == nil || activeUser.UID == 0 {
		ResponseError(c, 401, "未登录百度账号")
		return
	}

	path := c.Query("path")
	if path == "" {
		ResponseError(c, 400, "缺少 path 参数")
		return
	}

	pcs := pcsconfig.Config.ActiveUserBaiduPCS()
	if pcs == nil {
		ResponseError(c, 500, "获取网盘实例失败")
		return
	}

	info, err := pcs.LocateDownload(path)
	if err != nil {
		ResponseError(c, 500, "获取下载链接失败: "+err.Error())
		return
	}

	links := info.URLStrings(pcsconfig.Config.EnableHTTPS)
	if len(links) == 0 {
		ResponseError(c, 404, "未找到可用的下载链接")
		return
	}

	downloadLinks := make([]DownloadLinkInfo, 0, len(links))
	for _, link := range links {
		downloadLinks = append(downloadLinks, DownloadLinkInfo{
			URL:     link.String(),
			Expired: 0,
		})
	}

	ResponseSuccess(c, gin.H{
		"path":  path,
		"links": downloadLinks,
	})
}

type BatchDownloadRequest struct {
	Paths []string `json:"paths" form:"paths"`
}

type BatchDownloadResult struct {
	Path  string             `json:"path"`
	Links []DownloadLinkInfo `json:"links"`
	Error string             `json:"error,omitempty"`
}

func BatchGetDownloadLinks(c *gin.Context) {
	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser == nil || activeUser.UID == 0 {
		ResponseError(c, 401, "未登录百度账号")
		return
	}

	var req BatchDownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, 400, "参数错误")
		return
	}

	if len(req.Paths) == 0 {
		ResponseError(c, 400, "缺少 paths 参数")
		return
	}

	pcs := pcsconfig.Config.ActiveUserBaiduPCS()
	if pcs == nil {
		ResponseError(c, 500, "获取网盘实例失败")
		return
	}

	results := make([]BatchDownloadResult, 0, len(req.Paths))
	for _, path := range req.Paths {
		result := BatchDownloadResult{Path: path}

		info, err := pcs.LocateDownload(path)
		if err != nil {
			result.Error = err.Error()
			results = append(results, result)
			continue
		}

		links := info.URLStrings(pcsconfig.Config.EnableHTTPS)
		if len(links) == 0 {
			result.Error = "未找到可用的下载链接"
			results = append(results, result)
			continue
		}

		for _, link := range links {
			result.Links = append(result.Links, DownloadLinkInfo{
				URL:     link.String(),
				Expired: 0,
			})
		}
		results = append(results, result)
	}

	ResponseSuccess(c, gin.H{
		"results": results,
	})
}

func DownloadFile(c *gin.Context) {
	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser == nil || activeUser.UID == 0 {
		c.String(401, "未登录百度账号")
		return
	}

	filePath := c.Query("path")
	if filePath == "" {
		c.String(400, "缺少 path 参数")
		return
	}

	pcs := pcsconfig.Config.ActiveUserBaiduPCS()
	if pcs == nil {
		c.String(500, "获取网盘实例失败")
		return
	}

	info, pcsErr := pcs.LocateDownload(filePath)
	if pcsErr != nil {
		c.String(500, "获取下载链接失败: "+pcsErr.Error())
		return
	}

	links := info.URLStrings(pcsconfig.Config.EnableHTTPS)
	if len(links) == 0 {
		c.String(404, "未找到可用的下载链接")
		return
	}

	downloadURL := links[0].String()
	fileName := path.Base(filePath)
	fileName = url.PathEscape(fileName)

	client := &http.Client{}
	req, reqErr := http.NewRequest("GET", downloadURL, nil)
	if reqErr != nil {
		c.String(500, "创建请求失败: "+reqErr.Error())
		return
	}

	req.Header.Set("User-Agent", "netdisk")
	bduss := activeUser.BDUSS
	if bduss != "" {
		req.AddCookie(&http.Cookie{Name: "BDUSS", Value: bduss})
	}

	resp, respErr := client.Do(req)
	if respErr != nil {
		c.String(500, "下载文件失败: "+respErr.Error())
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusPartialContent {
		c.String(500, "服务器返回错误状态码: "+resp.Status)
		return
	}

	c.Header("Content-Disposition", "attachment; filename=\""+fileName+"\"; filename*=UTF-8''"+fileName)
	c.Header("Content-Type", "application/octet-stream")

	if resp.ContentLength > 0 {
		c.Header("Content-Length", strconv.FormatInt(resp.ContentLength, 10))
	}

	io.Copy(c.Writer, resp.Body)
}

func formatTime(timestamp int64) string {
	if timestamp == 0 {
		return "-"
	}
	return formatUnixTime(timestamp)
}
