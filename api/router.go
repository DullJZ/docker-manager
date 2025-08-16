package api

import (
	"net/http"

	"github.com/DullJZ/docker-manager/api/handlers"
	"github.com/DullJZ/docker-manager/auth"
	"github.com/DullJZ/docker-manager/service"
	"github.com/gin-gonic/gin"
)

// Router API路由器
type Router struct {
	engine           *gin.Engine
	imageHandler     *handlers.ImageHandler
	dockerHandler    *handlers.DockerHandler
	containerHandler *handlers.ContainerHandler
	networkHandler   *handlers.NetworkHandler
	fileHandler      *handlers.FileHandler
	sessionHandler   *handlers.SessionHandler
	composeHandler   *handlers.ComposeHandler
}

// NewRouter 创建新的路由器
func NewRouter(tokenManager *auth.TokenManager, dockerService *service.DockerService) *Router {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.SetTrustedProxies([]string{"localhost"})

	// 使用Token认证中间件
	r.Use(tokenManager.Middleware())

	// 创建服务实例
	imageService := service.NewImageService(dockerService)
	containerService := service.NewContainerService(dockerService)
	networkService := service.NewNetworkService(dockerService)
	fileService := service.NewFileService(dockerService)
	sessionService := service.NewSessionService(dockerService)
	composeService := service.NewComposeService(dockerService)

	// 创建处理器实例
	imageHandler := handlers.NewImageHandler(imageService)
	dockerHandler := handlers.NewDockerHandler(dockerService)
	containerHandler := handlers.NewContainerHandler(containerService)
	networkHandler := handlers.NewNetworkHandler(networkService)
	fileHandler := handlers.NewFileHandler(fileService)
	sessionHandler := handlers.NewSessionHandler(sessionService)
	composeHandler := handlers.NewComposeHandler(composeService)

	router := &Router{
		engine:           r,
		imageHandler:     imageHandler,
		dockerHandler:    dockerHandler,
		containerHandler: containerHandler,
		networkHandler:   networkHandler,
		fileHandler:      fileHandler,
		sessionHandler:   sessionHandler,
		composeHandler:   composeHandler,
	}

	router.setupRoutes()
	return router
}

// setupRoutes 设置路由
func (r *Router) setupRoutes() {
	// 健康检查
	r.engine.GET("/api", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// 镜像相关路由
	r.engine.POST("/api/pull_image", r.imageHandler.Pull)
	r.engine.POST("/api/delete_image", r.imageHandler.Delete)
	r.engine.POST("/api/list_images", r.imageHandler.List)
	r.engine.POST("/api/search_image", r.imageHandler.Search)

	// Docker相关路由
	r.engine.GET("/api/docker_info", r.dockerHandler.GetDockerInfo)
	r.engine.GET("/api/docker_version", r.dockerHandler.GetDockerVersion)

	// 容器相关路由
	r.engine.POST("/api/run_container", r.containerHandler.Run)
	r.engine.POST("/api/delete_container", r.containerHandler.Delete)
	r.engine.POST("/api/start_container", r.containerHandler.Start)
	r.engine.POST("/api/stop_container", r.containerHandler.Stop)
	r.engine.POST("/api/restart_container", r.containerHandler.Restart)
	r.engine.POST("/api/fetch_container_info", r.containerHandler.GetInfo)
	r.engine.POST("/api/fetch_container_logs", r.containerHandler.GetLogs)
	r.engine.POST("/api/search_container", r.containerHandler.Search)
	r.engine.GET("/api/get_running_containers", r.containerHandler.GetRunning)
	r.engine.GET("/api/get_all_containers", r.containerHandler.GetAll)
	r.engine.POST("/api/update_container", r.containerHandler.Update)

	// 网络相关路由
	r.engine.POST("/api/create_network", r.networkHandler.Create)
	r.engine.POST("/api/delete_network", r.networkHandler.Delete)
	r.engine.GET("/api/list_networks", r.networkHandler.List)
	r.engine.POST("/api/network_info", r.networkHandler.GetInfo)
	r.engine.POST("/api/connect_network", r.networkHandler.Connect)
	r.engine.POST("/api/disconnect_network", r.networkHandler.Disconnect)

	// 文件操作路由
	r.engine.POST("/api/copy_from_container", r.fileHandler.CopyFromContainer)
	r.engine.POST("/api/copy_to_container", r.fileHandler.CopyToContainer)

	// 交互式会话路由
	r.engine.POST("/api/create_exec_session", r.sessionHandler.CreateSession)
	r.engine.POST("/api/execute_command_in_session", r.sessionHandler.ExecuteCommand)
	r.engine.POST("/api/get_more_session_output", r.sessionHandler.GetMoreOutput)
	r.engine.POST("/api/close_exec_session", r.sessionHandler.CloseSession)

	// Compose支持路由
	r.engine.POST("/api/run_container_by_compose", r.composeHandler.RunByCompose)
}

// Run 启动服务器
func (r *Router) Run(addr string) error {
	return r.engine.Run(addr)
}
