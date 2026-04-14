package handler

import (
	"github.com/gin-gonic/gin"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/pcsconfig"
)

type LoginRequest struct {
	BDUSS   string `json:"bduss" form:"bduss"`
	PToken  string `json:"ptoken" form:"ptoken"`
	SToken  string `json:"stoken" form:"stoken"`
	Cookies string `json:"cookies" form:"cookies"`
}

type UserInfo struct {
	UID      uint64 `json:"uid"`
	Name     string `json:"name"`
	Sex      string `json:"sex"`
	Age      int    `json:"age"`
	Workdir  string `json:"workdir"`
	IsActive bool   `json:"is_active"`
}

func Login(c *gin.Context) {
	var req LoginRequest
	c.ShouldBind(&req)

	if req.BDUSS == "" && req.Cookies == "" {
		ResponseError(c, 400, "缺少 BDUSS 或 Cookies")
		return
	}

	baidu, err := pcsconfig.Config.SetupUserByBDUSS(req.BDUSS, req.PToken, req.SToken, req.Cookies)
	if err != nil {
		ResponseError(c, 401, "登录失败: "+err.Error())
		return
	}

	ResponseSuccess(c, gin.H{
		"uid":  baidu.UID,
		"name": baidu.Name,
		"sex":  baidu.Sex,
		"age":  int(baidu.Age),
	})
}

func GetUsers(c *gin.Context) {
	users := make([]UserInfo, 0, len(pcsconfig.Config.BaiduUserList))
	activeUID := pcsconfig.Config.BaiduActiveUID

	for _, u := range pcsconfig.Config.BaiduUserList {
		if u == nil {
			continue
		}
		users = append(users, UserInfo{
			UID:      u.UID,
			Name:     u.Name,
			Sex:      u.Sex,
			Age:      int(u.Age),
			Workdir:  u.Workdir,
			IsActive: u.UID == activeUID,
		})
	}

	ResponseSuccess(c, gin.H{
		"users":      users,
		"active_uid": activeUID,
	})
}

func SwitchUser(c *gin.Context) {
	var req struct {
		UID uint64 `json:"uid" form:"uid"`
	}
	c.ShouldBind(&req)

	if req.UID == 0 {
		ResponseError(c, 400, "缺少 uid 参数")
		return
	}

	user, err := pcsconfig.Config.SwitchUser(&pcsconfig.BaiduBase{
		UID: req.UID,
	})
	if err != nil {
		ResponseError(c, 404, "用户不存在")
		return
	}

	ResponseSuccess(c, gin.H{
		"uid":  user.UID,
		"name": user.Name,
	})
}

func SetDefaultUser(c *gin.Context) {
	var req struct {
		UID uint64 `json:"uid" form:"uid"`
	}
	c.ShouldBind(&req)

	if req.UID == 0 {
		ResponseError(c, 400, "缺少 uid 参数")
		return
	}

	user, err := pcsconfig.Config.SwitchUser(&pcsconfig.BaiduBase{
		UID: req.UID,
	})
	if err != nil {
		ResponseError(c, 404, "用户不存在")
		return
	}

	pcsconfig.Config.BaiduActiveUID = user.UID
	pcsconfig.Config.Save()

	ResponseSuccess(c, gin.H{
		"uid":  user.UID,
		"name": user.Name,
	})
}

func GetServerConfig(c *gin.Context) {
	config, err := LoadServerConfig()
	if err != nil {
		ResponseError(c, 500, "获取配置失败")
		return
	}

	ResponseSuccess(c, gin.H{
		"enable_web": config.EnableWeb,
		"enable_api": config.EnableAPI,
		"web_port":   config.WebPort,
	})
}

func UpdateServerConfig(c *gin.Context) {
	var req struct {
		EnableWeb *bool `json:"enable_web"`
		EnableAPI *bool `json:"enable_api"`
		WebPort   *int  `json:"web_port"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseError(c, 400, "参数错误")
		return
	}

	config, err := LoadServerConfig()
	if err != nil {
		ResponseError(c, 500, "获取配置失败")
		return
	}

	if req.EnableWeb != nil {
		config.EnableWeb = *req.EnableWeb
	}
	if req.EnableAPI != nil {
		config.EnableAPI = *req.EnableAPI
	}
	if req.WebPort != nil {
		config.WebPort = *req.WebPort
	}

	if err := SaveServerConfig(config); err != nil {
		ResponseError(c, 500, "保存配置失败")
		return
	}

	ResponseSuccess(c, gin.H{
		"enable_web": config.EnableWeb,
		"enable_api": config.EnableAPI,
		"web_port":   config.WebPort,
	})
}
