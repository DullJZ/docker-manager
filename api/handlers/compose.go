package handlers

import (
	"net/http"

	"github.com/DullJZ/docker-manager/models"
	"github.com/DullJZ/docker-manager/service"
	"github.com/gin-gonic/gin"
)

// ComposeHandler Compose处理器
type ComposeHandler struct {
	composeService *service.ComposeService
}

// NewComposeHandler 创建Compose处理器
func NewComposeHandler(composeService *service.ComposeService) *ComposeHandler {
	return &ComposeHandler{
		composeService: composeService,
	}
}

// RunByCompose 通过docker-compose运行容器
func (ch *ComposeHandler) RunByCompose(c *gin.Context) {
	var req models.ComposeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result := ch.composeService.RunByCompose(req.ComposeFile, req.OtherFiles, req.DeleteTempDir)
	c.JSON(http.StatusOK, result)
}
