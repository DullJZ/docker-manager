package handlers

import (
	"net/http"

	"github.com/DullJZ/docker-manager/models"
	"github.com/DullJZ/docker-manager/service"
	"github.com/gin-gonic/gin"
)

// SessionHandler 会话处理器
type SessionHandler struct {
	sessionService *service.SessionService
}

// NewSessionHandler 创建会话处理器
func NewSessionHandler(sessionService *service.SessionService) *SessionHandler {
	return &SessionHandler{
		sessionService: sessionService,
	}
}

// CreateSession 创建交互式会话
func (sh *SessionHandler) CreateSession(c *gin.Context) {
	var req models.ContainerNameRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, err := sh.sessionService.CreateSession(req.ContainerName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	sessionID := session.GetSessionID()
	c.JSON(http.StatusOK, gin.H{"status": "success", "exec_session_id": sessionID})
}

// ExecuteCommand 在会话中执行命令
func (sh *SessionHandler) ExecuteCommand(c *gin.Context) {
	var req models.ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, exists := sh.sessionService.GetSession(req.ExecSessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if session.ContainerName != req.ContainerName {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container name"})
		return
	}

	output, err := session.ExecuteCommandWithTimeout(req.Cmd, 5)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "success", "output": output})
}

// GetMoreOutput 获取更多会话输出
func (sh *SessionHandler) GetMoreOutput(c *gin.Context) {
	var req models.ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, exists := sh.sessionService.GetSession(req.ExecSessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if session.ContainerName != req.ContainerName {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container name"})
		return
	}

	output, err := session.GetSessionOutput()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"output": output})
}

// CloseSession 关闭交互式会话
func (sh *SessionHandler) CloseSession(c *gin.Context) {
	var req models.ExecRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	session, exists := sh.sessionService.GetSession(req.ExecSessionID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "session not found"})
		return
	}
	if session.ContainerName != req.ContainerName {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid container name"})
		return
	}

	if sh.sessionService.CloseSession(req.ExecSessionID) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	} else {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to close session"})
	}
}
