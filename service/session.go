package service

import (
	"context"
	rawNet "net"
	"strings"
	"time"

	"github.com/DullJZ/docker-manager/utils"
	"github.com/docker/docker/api/types"
)

// ExecSession 交互式执行会话
type ExecSession struct {
	ContainerName string
	ExecID        string
	Hijack        types.HijackedResponse
	Cli           *DockerService
	Ctx           context.Context
}

// SessionService 会话服务
type SessionService struct {
	*DockerService
	sessions map[string]*ExecSession
}

// NewSessionService 创建会话服务
func NewSessionService(dockerService *DockerService) *SessionService {
	return &SessionService{
		DockerService: dockerService,
		sessions:      make(map[string]*ExecSession),
	}
}

// CreateSession 创建一个新的交互式执行会话
func (ss *SessionService) CreateSession(containerName string) (*ExecSession, error) {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/bash"},
		Tty:          true,
	}

	resp, err := ss.client.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return nil, err
	}

	hijack, err := ss.client.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return nil, err
	}

	session := &ExecSession{
		ContainerName: containerName,
		ExecID:        resp.ID,
		Hijack:        hijack,
		Cli:           ss.DockerService,
		Ctx:           ctx,
	}

	// 存储会话
	sessionID := session.ExecID
	if len(sessionID) > 12 {
		sessionID = sessionID[:12] // 截断 ExecID
	}
	ss.sessions[sessionID] = session

	return session, nil
}

// GetSession 获取会话
func (ss *SessionService) GetSession(sessionID string) (*ExecSession, bool) {
	session, exists := ss.sessions[sessionID]
	return session, exists
}

// ExecuteCommandWithTimeout 执行命令并等待指定时间后返回输出
func (s *ExecSession) ExecuteCommandWithTimeout(cmd string, timeout int) (string, error) {
	// 发送命令
	_, err := s.Hijack.Conn.Write([]byte(cmd + "\n"))
	if err != nil {
		return "", err
	}

	// 等待2s
	time.Sleep(2 * time.Second)
	counter := 1

	// 读取所有可用的输出
	var output strings.Builder
	buffer := make([]byte, 4096)

	for counter <= timeout {
		n, err := s.Hijack.Reader.Read(buffer)
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			// 如果没有更多数据可读，退出循环
			if strings.Contains(err.Error(), "i/o timeout") {
				break
			}
		}
		if n > 0 {
			output.Write(buffer[:n])
			break
		}
		// 如果没有数据可读，等待1s
		time.Sleep(time.Second)
		counter++
	}

	// 清理 ANSI 转义序列
	cleanedOutput := utils.CleanANSIEscapes(output.String())
	return cleanedOutput, nil
}

// GetSessionOutput 获取会话输出
func (s *ExecSession) GetSessionOutput() (string, error) {
	var output strings.Builder
	buffer := make([]byte, 4096)

	// 设置短暂的超时来检查是否有数据
	s.Hijack.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
	defer s.Hijack.Conn.SetReadDeadline(time.Time{})

	for {
		n, err := s.Hijack.Reader.Read(buffer)
		if err != nil {
			if netErr, ok := err.(rawNet.Error); ok && netErr.Timeout() {
				// 超时，没有更多数据
				break
			}
			return "", err
		}
		if n > 0 {
			output.Write(buffer[:n])
			// 重新设置超时以继续读取
			s.Hijack.Conn.SetReadDeadline(time.Now().Add(100 * time.Millisecond))
		} else {
			break
		}
	}

	cleanedOutput := utils.CleanANSIEscapes(output.String())
	return cleanedOutput, nil
}

// Close 关闭会话
func (s *ExecSession) Close() {
	// 发送exit命令
	s.Hijack.Conn.Write([]byte("exit\n"))
	// 关闭连接
	s.Hijack.Close()
}

// CloseSession 关闭并删除会话
func (ss *SessionService) CloseSession(sessionID string) bool {
	session, exists := ss.sessions[sessionID]
	if !exists {
		return false
	}

	session.Close()
	delete(ss.sessions, sessionID)
	return true
}

// GetSessionID 获取会话ID（截断的ExecID）
func (s *ExecSession) GetSessionID() string {
	if len(s.ExecID) > 12 {
		return s.ExecID[:12]
	}
	return s.ExecID
}
