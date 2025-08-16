package service

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"time"
)

// ComposeService Compose服务
type ComposeService struct {
	*DockerService
}

// NewComposeService 创建Compose服务
func NewComposeService(dockerService *DockerService) *ComposeService {
	return &ComposeService{DockerService: dockerService}
}

// RunByCompose 通过docker-compose运行容器
func (cs *ComposeService) RunByCompose(composeFile string, otherFiles map[string]string, deleteTempDir bool) map[string]interface{} {
	// 在当前目录下建立临时目录
	cwd, err := os.Getwd()
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": "get current dir failed", "raw_message": err.Error()}
	}
	tempDir := filepath.Join(cwd, "docker-compose-"+strconv.FormatInt(time.Now().Unix(), 10))
	os.MkdirAll(tempDir, os.ModePerm)

	// 将compose文件写入临时目录
	name := tempDir + "/docker-compose.yaml"
	os.WriteFile(name, []byte(composeFile), 0644)

	// 将其他文件写入临时目录
	for n, file := range otherFiles {
		saveName := filepath.Join(tempDir, n)
		os.WriteFile(saveName, []byte(file), 0644)
	}

	// 执行docker-compose up -d
	cmd := exec.Command("docker-compose", "-f", name, "up", "-d")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		// 尝试使用docker compose
		cmd = exec.Command("docker", "compose", "-f", name, "up", "-d")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			if deleteTempDir {
				os.RemoveAll(tempDir)
			}
			return map[string]interface{}{"status": "fail", "message": stderr.String()}
		}
		if deleteTempDir {
			os.RemoveAll(tempDir)
		}
		return map[string]interface{}{"status": "success", "message": stdout.String()}
	}

	// 删除临时目录
	if deleteTempDir {
		os.RemoveAll(tempDir)
	}
	return map[string]interface{}{"status": "success", "message": stdout.String()}
}
