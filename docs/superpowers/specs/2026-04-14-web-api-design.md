# BaiduPCS-Go Web/API 服务设计

## 概述

为 BaiduPCS-Go 添加 Web 界面和 REST API 服务，支持浏览百度网盘文件、获取下载链接、账号管理。

## 配置文件

### server.json
```json
{
  "enable_web": true,
  "enable_api": true,
  "web_port": 8080,
  "api_port": 8081
}
```

配置文件路径：与 `pcs_config.json` 同目录（`~/.config/BaiduPCS-Go/server.json` 或 Windows `%APPDATA%/BaiduPCS-Go/server.json`）

## 项目结构

```
internal/
  web/
    server.go           # 服务器入口
    router.go          # Gin 路由配置
    handler/
      files.go          # 文件操作 API
      auth.go           # 账号管理 API
    templates/
      index.html        # Web UI 页面
```

## Web UI 功能

| 功能 | 说明 |
|------|------|
| 文件浏览 | 目录列表、文件名、大小、修改时间 |
| 下载链接 | 点击文件获取直链 |
| 账号登录 | 输入 BDUSS/STOKEN 登录验证 |
| 账号管理 | 切换账号、设置默认账号 |

## API 接口

### 文件操作

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/files` | GET | `path` (可选, 默认"/") | 获取目录文件列表 |
| `/api/download` | GET | `path` | 获取文件下载链接 |

### 账号管理

| 端点 | 方法 | 参数 | 说明 |
|------|------|------|------|
| `/api/login` | POST | `bduss`, `ptoken`, `stoken` | BDUSS 登录验证 |
| `/api/users` | GET | - | 获取账号列表 |
| `/api/users/switch` | POST | `uid` | 切换账号 |
| `/api/users/default` | POST | `uid` | 设置默认账号 |

### API 响应格式

```json
{
  "code": 0,
  "message": "success",
  "data": {}
}
```

## 集成方式

- `main.go` 初始化时加载 `server.json`
- 如果 `enable_web` 或 `enable_api` 为 true，启动 Gin 服务器
- Web 和 API 可独立开启
- 使用现有 `pcsconfig.Config` 和 `pcsconfig.Config.ActiveUserBaiduPCS()`

## 技术选型

- Web 框架：Gin
- 前端：原生 HTML/CSS/JavaScript
- API：REST JSON

## 依赖

- `github.com/gin-gonic/gin v1.9.1`
