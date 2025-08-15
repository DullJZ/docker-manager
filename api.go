package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types/network"
	"github.com/gin-gonic/gin"
)

// 定义 token 白名单
var tokenWhitelist = make(map[string]struct{})

// 是否启用 token 校验
var tokenAuthEnabled = true

// 定义命令行参数
var (
	listenIP   string
	listenPort int
)

// 创建交互式TTY会话
type ExecSessionManager struct {
	Sessions map[string]*ExecSession
}

var sessionManager = ExecSessionManager{
	Sessions: make(map[string]*ExecSession),
}

// 从文件加载 token 白名单
func loadTokenWhitelist() error {
	// 获取当前执行文件所在目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	// 打开 tokens.txt 文件
	file, err := os.Open(filepath.Join(dir, "tokens.txt"))
	if err != nil {
		// 如果文件不存在，则关闭 token 校验，允许所有请求通过
		if os.IsNotExist(err) {
			tokenAuthEnabled = false
			return nil
		}
		return err
	}
	defer file.Close()

	// 清空现有白名单
	tokenWhitelist = make(map[string]struct{})

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 添加到白名单
		tokenWhitelist[line] = struct{}{}
	}

	tokenAuthEnabled = true // 文件存在且加载成功，启用校验
	return scanner.Err()
}

// Token 校验中间件
func tokenAuthMiddleware(c *gin.Context) {
	// 如果未启用 token 校验，直接放行
	if !tokenAuthEnabled {
		c.Next()
		return
	}
	// 如果为 GET /api，直接放行
	if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/api" {
		c.Next()
		return
	}
	token := c.GetHeader("Authorization")
	if _, ok := tokenWhitelist[token]; !ok {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or missing token", "status": "fail"})
		return
	}
	c.Next()
}

func main() {
	// 设置命令行参数
	flag.StringVar(&listenIP, "ip", "127.0.0.1", "监听IP地址")
	flag.IntVar(&listenPort, "port", 15000, "监听端口号")
	flag.Parse()

	// 验证参数
	if listenPort < 1 || listenPort > 65535 {
		fmt.Printf("无效的端口号: %d, 端口范围应为 1-65535\n", listenPort)
		os.Exit(1)
	}

	// 加载 token 白名单
	if err := loadTokenWhitelist(); err != nil {
		panic("Failed to load token whitelist: " + err.Error())
	}

	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.SetTrustedProxies([]string{"localhost"})

	r.Use(tokenAuthMiddleware)

	r.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	r.POST("/api/delete_image", func(c *gin.Context) {
		var data struct {
			ImageName string `json:"image_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "fail"})
			return
		}
		client := GetDockerClient("")
		c.JSON(http.StatusOK, DeleteImage(client, data.ImageName))
	})

	r.POST("/api/pull_image", func(c *gin.Context) {
		var data struct {
			Remote    string `json:"remote"`
			ImageName string `json:"image_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "fail"})
			return
		}
		client := GetDockerClient("")
		p := PullImage(client, data.ImageName)
		if p["status"] == "fail" {
			c.JSON(http.StatusBadRequest, p)
			return
		}
		if p["status"] == "success" {
			info, err := GetImageInfo(client, data.ImageName)
			if err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "fail"})
				return
			}
			c.JSON(http.StatusOK, gin.H{"status": "success", "info": info})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "unknown error", "status": "fail"})
	})

	r.POST("/api/list_images", func(c *gin.Context) {
		client := GetDockerClient("")
		images, err := ListImages(client)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, images)
	})

	r.POST("/api/restart_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		result := RestartContainer(client, data.ContainerName)

		c.JSON(http.StatusOK, result)
	})

	r.POST("/api/start_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		client := GetDockerClient("")
		result := StartContainer(client, data.ContainerName)

		c.JSON(http.StatusOK, result)
	})

	r.POST("/api/stop_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		client := GetDockerClient("")
		result := StopContainer(client, data.ContainerName)

		c.JSON(http.StatusOK, result)
	})

	r.GET("/api/get_docker_info", func(c *gin.Context) {
		client := GetDockerClient("")
		info, err := GetDockerInfo(client)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, info)
	})

	r.GET("/api/get_docker_version", func(c *gin.Context) {
		client := GetDockerClient("")
		version, err := GetDockerVersion(client)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, version)
	})

	r.GET("/api/get_running_containers", func(c *gin.Context) {
		client := GetDockerClient("")
		containers, err := GetRunningContainers(client)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, containers)
	})

	r.GET("/api/get_all_containers", func(c *gin.Context) {
		client := GetDockerClient("")
		containers, err := GetAllContainers(client)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, containers)
	})

	r.POST("/api/fetch_container_info", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		containerInfo, err := GetContainerInfo(client, data.ContainerName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, containerInfo)
	})

	r.POST("/api/search_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		containers := SearchContainers(client, data.ContainerName)
		c.JSON(http.StatusOK, containers)
	})

	r.POST("/api/search_image", func(c *gin.Context) {
		var data struct {
			ImageName string `json:"image_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		images := SearchImages(client, data.ImageName)
		c.JSON(http.StatusOK, images)
	})

	r.POST("/api/run_container", func(c *gin.Context) {
		var data struct {
			ImageName     string   `json:"image_name"`
			ContainerName string   `json:"container_name"`
			Cmd           string   `json:"cmd"`
			AddCaps       []string `json:"add_caps"`
			HostName      string   `json:"host_name"`
			PortMap       []string `json:"port_map"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		container, r := RunContainer(client, data.ImageName, data.ContainerName, data.Cmd, data.AddCaps, data.HostName, data.PortMap)
		if r["status"] == "fail" {
			c.JSON(http.StatusBadRequest, r)
			return
		}
		c.JSON(http.StatusOK, container)
	})

	r.POST("/api/delete_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		c.JSON(http.StatusOK, DeleteContainer(client, data.ContainerName))
	})

	r.POST("/api/fetch_container_logs", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		logs, err := GetContainerLogs(client, data.ContainerName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"data": logs})
	})

	r.POST("/api/update_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
			Memery        int64  `json:"memery"`
			CpuShares     int64  `json:"cpuShares"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		response := UpdateContainer(client, data.ContainerName, data.Memery, data.CpuShares)
		if response["status"] == "fail" {
			c.JSON(http.StatusBadRequest, response)
			return
		}
		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/copy_from_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
			SourcePath    string `json:"source_path"`
			DestPath      string `json:"dest_path"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response := CopyFromContainer(data.ContainerName, data.SourcePath, data.DestPath)
		if response["status"] == "fail" {
			c.JSON(http.StatusBadRequest, response)
			return
		}
		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/copy_to_container", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
			SourcePath    string `json:"source_path"`
			DestPath      string `json:"dest_path"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response := CopyToContainer(data.ContainerName, data.SourcePath, data.DestPath)
		if response["status"] == "fail" {
			c.JSON(http.StatusBadRequest, response)
			return
		}
		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/docker_exec", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
			Cmd           string `json:"cmd"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		response, err := DockerExec(client, data.ContainerName, data.Cmd)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/run_container_by_compose", func(c *gin.Context) {
		var data struct {
			ComposeFile   string            `json:"compose_file"`
			OtherFiles    map[string]string `json:"other_files"`
			DeleteTempDir bool              `json:"delete_temp_dir"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		response := RunContainerByCompose(data.ComposeFile, data.OtherFiles, data.DeleteTempDir)
		c.JSON(http.StatusOK, response)
	})

	// 创建一个新的交互式会话
	r.POST("/api/create_exec_session", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		session, err := CreateExecSession(client, data.ContainerName)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		sessionID := session.ExecID
		if len(sessionID) > 12 {
			sessionID = sessionID[:12] // 截断 ExecID
		}
		sessionManager.Sessions[sessionID] = session
		c.JSON(http.StatusOK, gin.H{"status": "success", "exec_session_id": sessionID})
	})

	// 执行单个命令
	r.POST("/api/execute_command_in_session", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
			Cmd           string `json:"cmd"`
			ExecSessionID string `json:"exec_session_id"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		session, exists := sessionManager.Sessions[data.ExecSessionID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		if session.ContainerName != data.ContainerName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container name"})
			return
		}

		output, err := session.ExecuteCommandWithTimeout(data.Cmd, 5)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "success", "output": output})
	})

	// 读取更多会话输出
	r.POST("/api/get_more_session_output", func(c *gin.Context) {
		var data struct {
			ExecSessionID string `json:"exec_session_id"`
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		session, exists := sessionManager.Sessions[data.ExecSessionID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		if session.ContainerName != data.ContainerName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container name"})
			return
		}
		output, err := session.GetSessionOutput()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"output": output})
	})

	// 关闭交互式会话
	r.POST("/api/close_exec_session", func(c *gin.Context) {
		var data struct {
			ContainerName string `json:"container_name"`
			ExecSessionID string `json:"exec_session_id"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		session, exists := sessionManager.Sessions[data.ExecSessionID]
		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
			return
		}
		if session.ContainerName != data.ContainerName {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container name"})
			return
		}

		session.Close()
		delete(sessionManager.Sessions, data.ExecSessionID)

		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	r.POST("/api/create_network", func(c *gin.Context) {
		var data struct {
			NetworkName string `json:"network_name"`
			IsInternal  bool   `json:"is_internal"`
			Subnet      string `json:"subnet"` // 可选参数
			IPRange     string `json:"ip_range"` // 可选参数
			Gateway     string `json:"gateway"`   // 可选参数
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		ipamConfig := []network.IPAMConfig{
			{
				Subnet:  data.Subnet,
				IPRange: data.IPRange,
				Gateway: data.Gateway,
			},
		}
		response := CreateNetwork(client, data.NetworkName, data.IsInternal, ipamConfig)
		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/delete_network", func(c *gin.Context) {
		var data struct {
			NetworkName string `json:"network_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		response := DeleteNetwork(client, data.NetworkName)
		c.JSON(http.StatusOK, response)
	})

	r.GET("/api/list_networks", func(c *gin.Context) {
		client := GetDockerClient("")
		networks, err := ListNetworks(client)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, networks)
	})

	r.POST("/api/network_info", func(c *gin.Context) {
		var data struct {
			NetworkName string `json:"network_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		network, err := NetworkInfo(client, data.NetworkName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, network)
	})

	r.POST("/api/connect_network", func(c *gin.Context) {
		var data struct {
			NetworkName   string `json:"network_name"`
			ContainerName string `json:"container_name"`
			IPv4Address   string `json:"ipv4_address"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		response := ConnectNetwork(client, data.NetworkName, data.ContainerName, data.IPv4Address)
		c.JSON(http.StatusOK, response)
	})

	r.POST("/api/disconnect_network", func(c *gin.Context) {
		var data struct {
			NetworkName   string `json:"network_name"`
			ContainerName string `json:"container_name"`
		}
		if err := c.ShouldBindJSON(&data); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		client := GetDockerClient("")
		response := DisconnectNetwork(client, data.NetworkName, data.ContainerName)
		c.JSON(http.StatusOK, response)
	})

	// 使用命令行参数指定的地址启动服务
	addr := fmt.Sprintf("%s:%d", listenIP, listenPort)
	r.Run(addr)
}
