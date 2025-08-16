package handlers

import (
	"github.com/DullJZ/docker-manager/service"
	"github.com/gin-gonic/gin"
)

// DockerHandler Docker相关请求处理
type DockerHandler struct {
	dockerService *service.DockerService
}

// NewDockerHandler 创建DockerHandler
func NewDockerHandler(dockerService *service.DockerService) *DockerHandler {
	return &DockerHandler{
		dockerService: dockerService,
	}
}

func (h *DockerHandler) GetDockerInfo(c *gin.Context) {
	info, err := h.dockerService.GetDockerInfo()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, info)
}

func (h *DockerHandler) GetDockerVersion(c *gin.Context) {
	version, err := h.dockerService.GetDockerVersion()
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, version)
}