package handlers

import (
	"net/http"

	"github.com/DullJZ/docker-manager/models"
	"github.com/DullJZ/docker-manager/service"
	"github.com/gin-gonic/gin"
)

// ImageHandler 镜像处理器
type ImageHandler struct {
	imageService *service.ImageService
}

// NewImageHandler 创建镜像处理器
func NewImageHandler(imageService *service.ImageService) *ImageHandler {
	return &ImageHandler{
		imageService: imageService,
	}
}

// Pull 拉取镜像
func (ih *ImageHandler) Pull(c *gin.Context) {
	var req models.ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "fail"})
		return
	}

	result := ih.imageService.Pull(req.ImageName)
	if result["status"] == "fail" {
		c.JSON(http.StatusBadRequest, result)
		return
	}

	if result["status"] == "success" {
		info, err := ih.imageService.GetInfo(req.ImageName)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "fail"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "success", "info": info})
		return
	}

	c.JSON(http.StatusBadRequest, gin.H{"error": "unknown error", "status": "fail"})
}

// Delete 删除镜像
func (ih *ImageHandler) Delete(c *gin.Context) {
	var req models.ImageRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": "fail"})
		return
	}

	result := ih.imageService.Delete(req.ImageName)
	c.JSON(http.StatusOK, result)
}

// List 列出镜像
func (ih *ImageHandler) List(c *gin.Context) {
	images, err := ih.imageService.List()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, images)
}

// Search 搜索镜像
func (ih *ImageHandler) Search(c *gin.Context) {
	var req models.SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	images := ih.imageService.Search(req.Query)
	c.JSON(http.StatusOK, images)
}
