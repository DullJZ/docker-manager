package handlers

import (
	"net/http"

	"github.com/DullJZ/docker-manager/models"
	"github.com/DullJZ/docker-manager/service"
	"github.com/docker/docker/api/types/network"
	"github.com/gin-gonic/gin"
)

// NetworkHandler 网络处理器
type NetworkHandler struct {
	networkService *service.NetworkService
}

// NewNetworkHandler 创建网络处理器
func NewNetworkHandler(networkService *service.NetworkService) *NetworkHandler {
	return &NetworkHandler{
		networkService: networkService,
	}
}

// Create 创建网络
func (nh *NetworkHandler) Create(c *gin.Context) {
	var req models.NetworkRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ipamConfig := []network.IPAMConfig{
		{
			Subnet:  req.Subnet,
			IPRange: req.IPRange,
			Gateway: req.Gateway,
		},
	}

	result := nh.networkService.Create(req.NetworkName, req.IsInternal, ipamConfig)
	c.JSON(http.StatusOK, result)
}

// Delete 删除网络
func (nh *NetworkHandler) Delete(c *gin.Context) {
	var req models.NetworkNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := nh.networkService.Delete(req.NetworkName)
	c.JSON(http.StatusOK, result)
}

// List 列出网络
func (nh *NetworkHandler) List(c *gin.Context) {
	networks, err := nh.networkService.List()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, networks)
}

// GetInfo 获取网络信息
func (nh *NetworkHandler) GetInfo(c *gin.Context) {
	var req models.NetworkNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	network, err := nh.networkService.GetInfo(req.NetworkName)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, network)
}

// Connect 连接容器到网络
func (nh *NetworkHandler) Connect(c *gin.Context) {
	var req models.NetworkConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := nh.networkService.Connect(req.NetworkName, req.ContainerName, req.IPv4Address)
	c.JSON(http.StatusOK, result)
}

// Disconnect 断开容器与网络连接
func (nh *NetworkHandler) Disconnect(c *gin.Context) {
	var req models.NetworkConnectRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := nh.networkService.Disconnect(req.NetworkName, req.ContainerName)
	c.JSON(http.StatusOK, result)
}
