package web

import (
	"fmt"
	"sync"

	"github.com/qjfoidnh/BaiduPCS-Go/internal/web/handler"
)

type WebServer struct {
	config *handler.ServerConfig
	wg     sync.WaitGroup
}

var (
	server     *WebServer
	serverOnce sync.Once
)

func StartWebServer() error {
	config, err := handler.LoadServerConfig()
	if err != nil {
		return fmt.Errorf("加载服务器配置失败: %v", err)
	}

	if !config.EnableWeb && !config.EnableAPI {
		return nil
	}

	serverOnce.Do(func() {
		server = &WebServer{
			config: config,
		}

		ginWeb := SetupWebRouter()
		server.wg.Add(1)
		go func() {
			defer server.wg.Done()
			addr := fmt.Sprintf(":%d", config.WebPort)
			fmt.Printf("Web/API 服务器启动中: http://localhost%s\n", addr)
			if err := ginWeb.Run(addr); err != nil {
				fmt.Printf("服务器错误: %v\n", err)
			}
		}()
	})

	return nil
}

func StopWebServer() {
	if server != nil {
		fmt.Println("正在停止 Web/API 服务器...")
	}
}
