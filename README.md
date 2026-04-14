# BaiduPanLinker

[English](#english) | [中文](#中文)

---

## English

### BaiduPanLinker - No-Tampermonkey Baidu Netdisk Direct Link Downloader

![Preview](./demo.png)

A web-based Baidu Netdisk management tool that allows you to browse files, get direct download links, and send files to Aria2 RPC downloader - all without installing any browser extensions like Tampermonkey.

### Features

- **No Browser Extensions Required** - Works directly in any modern browser
- **File Management** - Browse directories, view file info (size, modified time)
- **Direct Link Acquisition** - One-click get download links for files
- **Batch Operations** - Select multiple files and get links in batch
- **Aria2 Integration** - Send files directly to Aria2 RPC downloader
- **Multi-Account Support** - Manage multiple Baidu accounts
- **Copy Download Commands** - Copy as Aria2c or cUrl commands

### Quick Start

1. **Create `server.json` config:**
```json
{
    "enable_web": true,
    "web_port": 8080
}
```

2. **Start the program and visit:** http://localhost:8080

3. **Add your Baidu account:**
   - Click "Add Account" (添加账号)
   - Enter your BDUSS from Baidu Netdisk
   - Click Login

4. **Start downloading:**
   - Browse to your files
   - Select files with checkboxes
   - Click "Send to RPC" to download with Aria2
   - Or click "Copy Link" for direct links

### API Endpoints

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/api/files` | GET | List files in directory |
| `/api/files/download` | GET | Get download link |
| `/api/files/batch` | POST | Batch get download links |
| `/api/login` | POST | Add account |
| `/api/users` | GET | List accounts |
| `/api/users/switch` | POST | Switch account |

### Tech Stack

- **Backend**: Go + Gin
- **Frontend**: Pure HTML/CSS/JavaScript
- **Downloader**: Aria2 + AriaNg

### Related Projects

- [BaiduPCS-Go](https://github.com/qjfoidnh/BaiduPCS-Go) - Command-line version

### License

Apache License 2.0 - See [LICENSE](./LICENSE) for details.

---

## 中文

### 百度网盘免油猴直链下载工具

一款基于 Web 界面的百度网盘管理工具，无需安装任何浏览器插件（如油猴脚本），即可在浏览器中管理百度网盘文件、获取直链下载地址，并支持将文件发送到 Aria2 下载器。

### 功能特点

- **免油猴** - 无需安装任何浏览器插件，直接在浏览器中使用
- **文件管理** - 浏览器端浏览网盘目录，查看文件大小、修改时间
- **直链获取** - 一键获取文件直链下载地址
- **批量操作** - 勾选多个文件，批量获取下载链接
- **Aria2 集成** - 直接推送文件到 Aria2 RPC 下载器
- **多账号支持** - 管理多个百度账号，一键切换
- **复制命令** - 复制为 Aria2c 或 cUrl 下载命令

### 快速开始

1. **创建配置文件 `server.json`：**
```json
{
    "enable_web": true,
    "web_port": 8080
}
```

2. **启动程序后访问：** http://localhost:8080

3. **添加账号：**
   - 点击「添加账号」按钮
   - 输入从百度网盘获取的 BDUSS
   - 点击登录

4. **开始下载：**
   - 浏览到需要下载的文件
   - 勾选要下载的文件
   - 点击「发送到 RPC」推送到 Aria2 下载
   - 或点击「复制链接」获取直链

### API 接口

| 接口 | 方法 | 说明 |
|------|------|------|
| `/api/files` | GET | 获取文件列表 |
| `/api/files/download` | GET | 获取下载链接 |
| `/api/files/batch` | POST | 批量获取下载链接 |
| `/api/login` | POST | 添加账号 |
| `/api/users` | GET | 获取账号列表 |
| `/api/users/switch` | POST | 切换账号 |

### 技术栈

- **后端**：Go + Gin
- **前端**：原生 HTML/CSS/JavaScript
- **下载器**：Aria2 + AriaNg

### 相关项目

- [BaiduPCS-Go](https://github.com/qjfoidnh/BaiduPCS-Go) - 命令行版本

### 开源协议

Apache License 2.0 - 详见 [LICENSE](./LICENSE)
