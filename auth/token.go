package auth

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
)

// TokenManager Token管理器
type TokenManager struct {
	whitelist map[string]struct{}
	enabled   bool
}

// NewTokenManager 创建Token管理器
func NewTokenManager() *TokenManager {
	return &TokenManager{
		whitelist: make(map[string]struct{}),
		enabled:   false,
	}
}

// LoadFromFile 从文件加载token白名单
func (tm *TokenManager) LoadFromFile(filename string) error {
	// 获取当前执行文件所在目录
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		return err
	}

	// 打开 tokens.txt 文件
	file, err := os.Open(filepath.Join(dir, filename))
	if err != nil {
		// 如果文件不存在，则关闭 token 校验
		if os.IsNotExist(err) {
			tm.enabled = false
			return nil
		}
		return err
	}
	defer file.Close()

	// 清空现有白名单
	tm.whitelist = make(map[string]struct{})

	// 逐行读取文件
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// 跳过空行和注释行
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 添加到白名单
		tm.whitelist[line] = struct{}{}
	}

	tm.enabled = true // 文件存在且加载成功，启用校验
	return scanner.Err()
}

// IsValidToken 验证token是否有效
func (tm *TokenManager) IsValidToken(token string) bool {
	if !tm.enabled {
		return true
	}
	_, ok := tm.whitelist[token]
	return ok
}

// Middleware Token校验中间件
func (tm *TokenManager) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 如果未启用 token 校验，直接放行
		if !tm.enabled {
			c.Next()
			return
		}

		// 如果为 GET /api，直接放行
		if c.Request.Method == http.MethodGet && c.Request.URL.Path == "/api" {
			c.Next()
			return
		}

		token := c.GetHeader("Authorization")
		if !tm.IsValidToken(token) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":  "invalid or missing token",
				"status": "fail",
			})
			return
		}
		c.Next()
	}
}
