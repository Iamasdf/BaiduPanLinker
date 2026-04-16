package web

import (
	"github.com/gin-gonic/gin"
	"github.com/qjfoidnh/BaiduPCS-Go/internal/web/handler"
)

func SetupAPIRouter() *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	api := r.Group("/api")
	{
		files := api.Group("/files")
		{
			files.GET("", handler.GetFiles)
			files.GET("/download", handler.GetDownloadLink)
			files.GET("/download-file", handler.DownloadFile)
			files.POST("/download", handler.CreateDownload)
			files.POST("/batch", handler.BatchGetDownloadLinks)
		}

		download := api.Group("/download")
		{
			download.GET("/list", handler.ListDownloads)
			download.GET("/:task_id", handler.GetDownloadStatus)
			download.POST("/:task_id/pause", handler.PauseDownload)
			download.POST("/:task_id/resume", handler.ResumeDownload)
			download.POST("/:task_id/cancel", handler.CancelDownload)
			download.GET("/history", handler.GetDownloadHistory)
			download.POST("/clear-history", handler.ClearDownloadHistory)
			download.DELETE("/history/:task_id", handler.DeleteDownloadHistory)
			download.POST("/run", handler.CreateDownloadRun)
		}

		api.POST("/login", handler.Login)

		users := api.Group("/users")
		{
			users.GET("", handler.GetUsers)
			users.POST("/switch", handler.SwitchUser)
			users.POST("/default", handler.SetDefaultUser)
		}

		api.GET("/config", handler.GetServerConfig)
		api.POST("/config", handler.UpdateServerConfig)
		api.GET("/config/defaults", handler.GetDefaultConfig)
	}

	return r
}

func SetupWebRouter() *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware())

	r.GET("/", handler.Index)

	api := r.Group("/api")
	{
		files := api.Group("/files")
		{
			files.GET("", handler.GetFiles)
			files.GET("/download", handler.GetDownloadLink)
			files.GET("/download-file", handler.DownloadFile)
			files.POST("/download", handler.CreateDownload)
			files.POST("/batch", handler.BatchGetDownloadLinks)
		}

		download := api.Group("/download")
		{
			download.GET("/list", handler.ListDownloads)
			download.GET("/:task_id", handler.GetDownloadStatus)
			download.POST("/:task_id/pause", handler.PauseDownload)
			download.POST("/:task_id/resume", handler.ResumeDownload)
			download.POST("/:task_id/cancel", handler.CancelDownload)
			download.GET("/history", handler.GetDownloadHistory)
			download.POST("/clear-history", handler.ClearDownloadHistory)
			download.DELETE("/history/:task_id", handler.DeleteDownloadHistory)
			download.POST("/run", handler.CreateDownloadRun)
		}

		api.POST("/login", handler.Login)

		users := api.Group("/users")
		{
			users.GET("", handler.GetUsers)
			users.POST("/switch", handler.SwitchUser)
			users.POST("/default", handler.SetDefaultUser)
		}

		api.GET("/config", handler.GetServerConfig)
		api.POST("/config", handler.UpdateServerConfig)
		api.GET("/config/defaults", handler.GetDefaultConfig)
	}

	return r
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
