package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcscommand"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
)

type DownloadRequest struct {
	Path    string `json:"path" form:"path"`
	SaveDir string `json:"save_dir" form:"save_dir"`
}

func CreateDownload(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.ShouldBind(&req)
	}

	if req.Path == "" {
		ResponseError(c, 400, "缺少 path 参数")
		return
	}

	if req.SaveDir == "" {
		ResponseError(c, 400, "缺少 save_dir 参数")
		return
	}

	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser == nil || activeUser.UID == 0 {
		ResponseError(c, 401, "未登录百度账号")
		return
	}

	task := downloadMgr.CreateTask(req.Path, req.SaveDir)

	userPCS := pcsconfig.Config.ActiveUserBaiduPCS()
	downloadPCS := userPCS

	runner := NewDownloadTaskRunnerWithPCS(task, userPCS, downloadPCS)

	RegisterDownloadRunner(task.ID, runner)
	runner.Run()

	ResponseSuccess(c, gin.H{
		"task_id":   task.ID,
		"path":      task.Path,
		"save_path": task.SavePath,
	})
}

func CreateDownloadRun(c *gin.Context) {
	var req DownloadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.ShouldBind(&req)
	}

	if req.Path == "" {
		ResponseError(c, 400, "缺少 path 参数")
		return
	}

	if req.SaveDir == "" {
		ResponseError(c, 400, "缺少 save_dir 参数")
		return
	}

	activeUser := pcsconfig.Config.ActiveUser()
	if activeUser == nil || activeUser.UID == 0 {
		ResponseError(c, 401, "未登录百度账号")
		return
	}

	task := downloadMgr.CreateTask(req.Path, req.SaveDir)
	taskID := task.ID

	// 使用登录用户的 PCS
	pcs := pcsconfig.Config.ActiveUserBaiduPCS()
	if pcs == nil {
		task.Status = "failed"
		task.Error = "获取网盘实例失败"
		ResponseError(c, 500, "获取网盘实例失败")
		return
	}

	// 异步执行 RunDownload（进度在终端打印，Web界面也显示）
	go func() {
		pcscommand.RunDownload([]string{req.Path}, &pcscommand.DownloadOptions{
			SaveTo:      req.SaveDir,
			IsOverwrite: true,
			PCS:         pcs,
			ProgressCallback: func(downloaded, total int64, speed string, timeLeft string) {
				downloadMgr.UpdateTask(taskID, func(t *WebDownloadTask) {
					t.Downloaded = downloaded
					t.Total = total
					if total > 0 {
						t.Progress = int(float64(downloaded) / float64(total) * 100)
					}
					t.Speed = speed
					t.TimeLeft = timeLeft
					t.Status = "running"
				})
			},
		})

		downloadMgr.UpdateTask(taskID, func(t *WebDownloadTask) {
			t.Status = "completed"
			t.Progress = 100
		})
		downloadMgr.RecordHistory(downloadMgr.GetTask(taskID))
	}()

	ResponseSuccess(c, gin.H{
		"task_id":   task.ID,
		"path":      task.Path,
		"save_path": task.SavePath,
	})
}

func GetDownloadStatus(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		ResponseError(c, 400, "缺少 task_id 参数")
		return
	}

	task := downloadMgr.GetTask(taskID)
	if task == nil {
		ResponseError(c, 404, "任务不存在")
		return
	}

	ResponseSuccess(c, task)
}

func ListDownloads(c *gin.Context) {
	tasks := downloadMgr.GetAllTasks()
	ResponseSuccess(c, gin.H{
		"tasks": tasks,
	})
}

func PauseDownload(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		ResponseError(c, 400, "缺少 task_id 参数")
		return
	}

	runner := GetDownloadRunner(taskID)
	if runner == nil {
		ResponseError(c, 404, "任务不存在")
		return
	}

	runner.Pause()
	ResponseSuccess(c, gin.H{
		"message": "已暂停",
	})
}

func ResumeDownload(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		ResponseError(c, 400, "缺少 task_id 参数")
		return
	}

	runner := GetDownloadRunner(taskID)
	if runner == nil {
		ResponseError(c, 404, "任务不存在")
		return
	}

	runner.Resume()
	ResponseSuccess(c, gin.H{
		"message": "已恢复",
	})
}

func CancelDownload(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		ResponseError(c, 400, "缺少 task_id 参数")
		return
	}

	runner := GetDownloadRunner(taskID)
	if runner != nil {
		runner.Cancel()
		RemoveDownloadRunner(taskID)
	}

	downloadMgr.RemoveTask(taskID)
	downloadMgr.DeleteHistory(taskID)

	ResponseSuccess(c, gin.H{
		"message": "已取消",
	})
}

func GetDownloadHistory(c *gin.Context) {
	history := downloadMgr.LoadHistory()
	ResponseSuccess(c, gin.H{
		"history": history,
	})
}

func ClearDownloadHistory(c *gin.Context) {
	downloadMgr.ClearHistory()
	ResponseSuccess(c, gin.H{
		"message": "历史记录已清除",
	})
}

func DeleteDownloadHistory(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		ResponseError(c, 400, "缺少 task_id 参数")
		return
	}

	success := downloadMgr.DeleteHistory(taskID)
	if success {
		ResponseSuccess(c, gin.H{
			"message": "删除成功",
		})
	} else {
		ResponseError(c, 404, "历史记录不存在")
	}
}
