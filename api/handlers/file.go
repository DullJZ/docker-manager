package handlers

import (
	"net/http"

	"github.com/DullJZ/docker-manager/models"
	"github.com/DullJZ/docker-manager/service"
	"github.com/gin-gonic/gin"
)

// FileHandler 文件操作处理器
type FileHandler struct {
	fileService *service.FileService
}

// NewFileHandler 创建文件操作处理器
func NewFileHandler(fileService *service.FileService) *FileHandler {
	return &FileHandler{
		fileService: fileService,
	}
}

// CopyFromContainer 从容器拷贝文件到主机
func (fh *FileHandler) CopyFromContainer(c *gin.Context) {
	var req models.FileOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := fh.fileService.CopyFromContainer(req.ContainerName, req.SourcePath, req.DestPath)
	if result["status"] == "fail" {
		c.JSON(http.StatusBadRequest, result)
		return
	}
	c.JSON(http.StatusOK, result)
}

// CopyToContainer 从主机拷贝文件到容器
func (fh *FileHandler) CopyToContainer(c *gin.Context) {
	var req models.FileOperationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := fh.fileService.CopyToContainer(req.ContainerName, req.SourcePath, req.DestPath)
	if result["status"] == "fail" {
		c.JSON(http.StatusBadRequest, result)
		return
	}
	c.JSON(http.StatusOK, result)
}
