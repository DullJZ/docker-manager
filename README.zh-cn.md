# Docker Manager

[English](README.md) | 简体中文

## 项目简介
Docker Manager 是一个基于 Go 语言开发的 Docker 管理 API 服务，提供丰富的容器、镜像、网络等管理接口，支持通过 HTTP API 对 Docker 进行自动化运维和管理。

## 功能特性
- 镜像管理：拉取、删除、搜索、查看镜像信息
- 容器管理：启动、停止、重启、删除、运行新容器、获取容器信息、获取日志、模糊搜索
- 网络管理：创建、删除、连接、断开网络，获取网络信息
- 交互式会话：支持创建交互式 TTY 会话，执行命令，获取输出，关闭会话
- 文件操作：容器与主机之间文件互拷
- 资源管理：容器资源（CPU/内存）动态调整
- Docker 信息查询：版本、系统信息、运行容器、全部容器列表
- Compose 支持：通过 docker-compose 文件批量部署容器
- Trivy 漏洞扫描支持（部分功能）
- Token 白名单鉴权（可选）

## 快速开始
1. 安装依赖（需本地已安装 Docker 环境）
2. 启动服务：
	```bash
	go run api.go funcs.go
	# 或编译后运行
	go build -o docker-manager api.go funcs.go
	./docker-manager --ip 127.0.0.1 --port 15000
	```
3. 可选：在同目录下创建 `tokens.txt` 文件，启用 Token 白名单鉴权。

## API 接口说明
所有接口均需以 POST/GET 方式访问，部分接口需携带 Authorization 头部（如启用 Token 校验）。

### 镜像相关
- `POST /api/pull_image` 拉取镜像
- `POST /api/delete_image` 删除镜像
- `POST /api/list_images` 镜像列表
- `POST /api/search_image` 镜像模糊搜索
- `GET /api/get_docker_info` Docker 信息
- `GET /api/get_docker_version` Docker 版本

### 容器相关
- `POST /api/run_container` 新建并运行容器
- `POST /api/delete_container` 删除容器
- `POST /api/start_container` 启动容器
- `POST /api/stop_container` 停止容器
- `POST /api/restart_container` 重启容器
- `POST /api/fetch_container_info` 获取容器详细信息
- `POST /api/fetch_container_logs` 获取容器日志
- `POST /api/search_container` 容器模糊搜索
- `GET /api/get_running_containers` 运行中容器列表
- `GET /api/get_all_containers` 所有容器列表
- `POST /api/update_container` 动态调整容器资源

### 网络相关
- `POST /api/create_network` 创建网络
- `POST /api/delete_network` 删除网络
- `GET /api/list_networks` 网络列表
- `POST /api/network_info` 网络详情
- `POST /api/connect_network` 容器连接网络
- `POST /api/disconnect_network` 容器断开网络

### 文件操作
- `POST /api/copy_from_container` 从容器拷贝文件到主机
- `POST /api/copy_to_container` 从主机拷贝文件到容器

### 交互式会话
- `POST /api/create_exec_session` 创建交互式会话
- `POST /api/execute_command_in_session` 在会话中执行命令
- `POST /api/get_more_session_output` 获取更多会话输出
- `POST /api/close_exec_session` 关闭会话

### Compose 支持
- `POST /api/run_container_by_compose` 通过 docker-compose 文件批量部署容器

## Token 白名单鉴权
在项目根目录下创建 `tokens.txt` 文件，每行一个 token（支持注释和空行）。如无该文件则关闭鉴权，所有请求均可访问。

## 依赖
- Go 1.18+
- Docker 环境
- 主要依赖库：gin、docker/docker、lithammer/fuzzysearch、gopsutil

## 参考
- [Gin Web 框架](https://gin-gonic.com/)
- [Docker Go SDK](https://pkg.go.dev/github.com/docker/docker)

