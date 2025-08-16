package handlers

import (
	"net/http"

	"github.com/DullJZ/docker-manager/models"
	"github.com/DullJZ/docker-manager/service"
	"github.com/gin-gonic/gin"
)

// ContainerHandler 容器处理器
type ContainerHandler struct {
	containerService *service.ContainerService
}

// NewContainerHandler 创建容器处理器
func NewContainerHandler(containerService *service.ContainerService) *ContainerHandler {
	return &ContainerHandler{
		containerService: containerService,
	}
}

// Start 启动容器
func (ch *ContainerHandler) Start(c *gin.Context) {
	var req models.ContainerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := ch.containerService.Start(req.ContainerName)
	c.JSON(http.StatusOK, result)
}

// Stop 停止容器
func (ch *ContainerHandler) Stop(c *gin.Context) {
	var req models.ContainerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := ch.containerService.Stop(req.ContainerName)
	c.JSON(http.StatusOK, result)
}

// Restart 重启容器
func (ch *ContainerHandler) Restart(c *gin.Context) {
	var req models.ContainerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := ch.containerService.Restart(req.ContainerName)
	c.JSON(http.StatusOK, result)
}

// GetInfo 获取容器信息
func (ch *ContainerHandler) GetInfo(c *gin.Context) {
	var req models.ContainerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containerInfo, err := ch.containerService.GetInfo(req.ContainerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containerInfo)
}

// GetRunning 获取运行中的容器
func (ch *ContainerHandler) GetRunning(c *gin.Context) {
	containers, err := ch.containerService.GetRunning()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

// GetAll 获取所有容器
func (ch *ContainerHandler) GetAll(c *gin.Context) {
	containers, err := ch.containerService.GetAll()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, containers)
}

// Search 搜索容器
func (ch *ContainerHandler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	containers := ch.containerService.Search(req.Query)
	c.JSON(http.StatusOK, containers)
}

// Run 运行新容器
func (ch *ContainerHandler) Run(c *gin.Context) {
	var req models.ContainerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	container, result := ch.containerService.Run(
		req.ImageName,
		req.ContainerName,
		req.Cmd,
		req.AddCaps,
		req.HostName,
		req.PortMap,
	)

	if result["status"] == "fail" {
		c.JSON(http.StatusBadRequest, result)
		return
	}
	c.JSON(http.StatusOK, container)
}

// Delete 删除容器
func (ch *ContainerHandler) Delete(c *gin.Context) {
	var req models.ContainerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := ch.containerService.Delete(req.ContainerName)
	c.JSON(http.StatusOK, result)
}

// GetLogs 获取容器日志
func (ch *ContainerHandler) GetLogs(c *gin.Context) {
	var req models.ContainerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	logs, err := ch.containerService.GetLogs(req.ContainerName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": logs})
}

// Update 更新容器资源
func (ch *ContainerHandler) Update(c *gin.Context) {
	var req models.ContainerUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := ch.containerService.Update(req.ContainerName, req.Memory, req.CPUShares)
	if result["status"] == "fail" {
		c.JSON(http.StatusBadRequest, result)
		return
	}
	c.JSON(http.StatusOK, result)
}
