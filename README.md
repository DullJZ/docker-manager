# Docker Manager

English | [简体中文](README.zh-cn.md)

## Project Overview
Docker Manager is a Docker management API service developed in Go, providing rich interfaces for managing containers, images, networks, and more. It supports automated operations and management of Docker via HTTP API.

## Features
- Image Management: Pull, delete, search, view image information
- Container Management: Start, stop, restart, delete, run new containers, get container info, fetch logs, fuzzy search
- Network Management: Create, delete, connect, disconnect networks, get network info
- Interactive Sessions: Create interactive TTY sessions, execute commands, get output, close sessions
- File Operations: Copy files between containers and host
- Resource Management: Dynamically adjust container resources (CPU/memory)
- Docker Info Query: Version, system info, running containers, all container list
- Compose Support: Batch deploy containers via docker-compose file
- Trivy Vulnerability Scan Support (partial)
- Token Whitelist Authentication (optional)

## Quick Start

Docker should be installed and running on your system, and the user running this service must have Docker permissions.

### Run from Binary
1. Download the latest binary from the [Release page](https://github.com/DullJZ/docker-manager/releases)
2. Grant execute permission:
	```bash
	chmod +x docker-manager
	```
3. Token authentication (optional):
	Create a `tokens.txt` file in the same directory to enable token whitelist authentication.
4. Switch to root or a user with Docker access, then start the service:
	```bash
	sudo su # switch to root
	./docker-manager --ip 127.0.0.1 --port 15000
	```

### Run from Source
1. Install dependencies (Docker environment required locally)
	```bash
	go mod tidy
	```
2. Token authentication (optional):
	Create a `tokens.txt` file in the same directory to enable token whitelist authentication.
3. Start the service:
	```bash
	go run api.go funcs.go
	# Or build and run
	go build -o docker-manager api.go funcs.go
	./docker-manager --ip 127.0.0.1 --port 15000
	```

## API Endpoints
All endpoints support POST/GET requests. Some require the Authorization header (if token validation is enabled).

### Image
- `POST /api/pull_image` Pull image
- `POST /api/delete_image` Delete image
- `POST /api/list_images` List images
- `POST /api/search_image` Fuzzy search images
- `GET /api/get_docker_info` Docker info
- `GET /api/get_docker_version` Docker version

### Container
- `POST /api/run_container` Create and run container
- `POST /api/delete_container` Delete container
- `POST /api/start_container` Start container
- `POST /api/stop_container` Stop container
- `POST /api/restart_container` Restart container
- `POST /api/fetch_container_info` Get container details
- `POST /api/fetch_container_logs` Get container logs
- `POST /api/search_container` Fuzzy search containers
- `GET /api/get_running_containers` List running containers
- `GET /api/get_all_containers` List all containers
- `POST /api/update_container` Dynamically adjust container resources

### Network
- `POST /api/create_network` Create network
- `POST /api/delete_network` Delete network
- `GET /api/list_networks` List networks
- `POST /api/network_info` Network details
- `POST /api/connect_network` Connect container to network
- `POST /api/disconnect_network` Disconnect container from network

### File Operations
- `POST /api/copy_from_container` Copy file from container to host
- `POST /api/copy_to_container` Copy file from host to container

### Interactive Session
- `POST /api/create_exec_session` Create interactive session
- `POST /api/execute_command_in_session` Execute command in session
- `POST /api/get_more_session_output` Get more session output
- `POST /api/close_exec_session` Close session

### Compose Support
- `POST /api/run_container_by_compose` Batch deploy containers via docker-compose file

## Token Whitelist Authentication
Create a `tokens.txt` file in the project root directory, one token per line (supports comments and blank lines). If the file does not exist, authentication is disabled and all requests are allowed.

## Dependencies
- Go 1.18+
- Docker environment
- Main libraries: gin, docker/docker, lithammer/fuzzysearch, gopsutil

## References
- [Gin Web Framework](https://gin-gonic.com/)
- [Docker Go SDK](https://pkg.go.dev/github.com/docker/docker)
