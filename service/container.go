package service

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/go-connections/nat"
	"github.com/lithammer/fuzzysearch/fuzzy"
)

// ContainerService 容器服务
type ContainerService struct {
	*DockerService
}

// NewContainerService 创建容器服务
func NewContainerService(dockerService *DockerService) *ContainerService {
	return &ContainerService{DockerService: dockerService}
}

// Start 启动容器
func (cs *ContainerService) Start(containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cs.client.ContainerStart(ctx, containerName, container.StartOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// Stop 停止容器
func (cs *ContainerService) Stop(containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cs.client.ContainerStop(ctx, containerName, container.StopOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// Restart 重启容器
func (cs *ContainerService) Restart(containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cs.client.ContainerRestart(ctx, containerName, container.StopOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// GetInfo 获取容器信息
func (cs *ContainerService) GetInfo(containerName string) (types.ContainerJSON, error) {
	ctx := context.Background()
	c, err := cs.client.ContainerInspect(ctx, containerName)
	if err != nil {
		return types.ContainerJSON{}, err
	}
	return c, nil
}

// GetRunning 获取运行中的容器
func (cs *ContainerService) GetRunning() ([]types.Container, error) {
	ctx := context.Background()
	containers, err := cs.client.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// GetAll 获取所有容器
func (cs *ContainerService) GetAll() ([]types.Container, error) {
	ctx := context.Background()
	containers, err := cs.client.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// Search 搜索容器
func (cs *ContainerService) Search(containerName string) []types.Container {
	allContainers, err := cs.GetAll()
	if err != nil {
		return []types.Container{}
	}

	var containers []types.Container
	for _, container := range allContainers {
		if len(container.Names) > 0 {
			// 模糊搜索
			if fuzzy.Match(containerName, container.Names[0]) {
				containers = append(containers, container)
			}
		}
	}
	return containers
}

// Delete 删除容器
func (cs *ContainerService) Delete(containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cs.client.ContainerRemove(ctx, containerName, container.RemoveOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": "Remove Container Error", "raw_message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// GetLogs 获取容器日志
func (cs *ContainerService) GetLogs(containerName string) (string, error) {
	ctx := context.Background()
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true}
	out, err := cs.client.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return "", err
	}
	defer out.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	return buf.String(), nil
}

// Update 更新容器资源
func (cs *ContainerService) Update(containerName string, memory int64, cpu int64) map[string]interface{} {
	ctx := context.Background()
	updateConfig := container.UpdateConfig{
		Resources: container.Resources{
			Memory:     memory,
			MemorySwap: memory * 2,
			CPUShares:  cpu,
		},
	}
	_, err := cs.client.ContainerUpdate(ctx, containerName, updateConfig)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// Run 运行新容器
func (cs *ContainerService) Run(imageName, containerName, cmd string, addCaps []string, hostname string, portMap []string) (types.ContainerJSON, map[string]interface{}) {
	ctx := context.Background()

	if containerName == "" {
		containerName = "container_" + strconv.FormatInt(time.Now().Unix(), 10)
	}
	if cmd == "" {
		cmd = "/bin/sh"
	}
	if imageName == "" {
		return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "imageName is empty"}
	}

	// 确保DefaultNet存在
	if err := cs.ensureDefaultNetwork(); err != nil {
		return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "create DefaultNet error", "raw_message": err.Error()}
	}

	// 检测hostname
	if hostname == "" {
		hostname = containerName
	}

	// 处理端口映射
	portBindings, exposedPorts, err := cs.processPortMapping(portMap)
	if err != nil {
		return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": err.Error()}
	}

	response, err := cs.client.ContainerCreate(ctx, &container.Config{
		Image:        imageName,
		Tty:          true,
		Cmd:          []string{cmd},
		Hostname:     hostname,
		ExposedPorts: exposedPorts,
		Labels:       map[string]string{},
	}, &container.HostConfig{
		PortBindings: portBindings,
		CapAdd:       addCaps,
	}, &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"DefaultNet": {}, // 自动加入默认网络
		},
	}, nil, containerName)

	if err != nil {
		return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "create container error", "raw_message": err.Error()}
	}

	// 启动容器
	cs.Start(response.ID)

	// 获取容器信息
	containerInfo, err := cs.client.ContainerInspect(ctx, response.ID)
	if err != nil {
		return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "inspect container error", "raw_message": err.Error()}
	}

	return containerInfo, map[string]interface{}{"status": "success"}
}

// ensureDefaultNetwork 确保默认网络存在
func (cs *ContainerService) ensureDefaultNetwork() error {
	ctx := context.Background()
	networks, err := cs.client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return err
	}

	for _, network := range networks {
		if network.Name == "DefaultNet" {
			return nil // 网络已存在
		}
	}

	// 创建默认网络
	_, err = cs.client.NetworkCreate(ctx, "DefaultNet", types.NetworkCreate{})
	return err
}

// processPortMapping 处理端口映射
func (cs *ContainerService) processPortMapping(portMap []string) (nat.PortMap, nat.PortSet, error) {
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}

	for _, p := range portMap {
		if strings.Count(p, ":") != 1 {
			return nil, nil, fmt.Errorf("portMap error")
		}

		parts := strings.Split(p, ":")
		hostPort := parts[0]
		containerPort := parts[1]

		portBindings[nat.Port(containerPort+"/tcp")] = []nat.PortBinding{
			{
				HostIP:   "0.0.0.0",
				HostPort: hostPort,
			},
		}
		exposedPorts[nat.Port(containerPort+"/tcp")] = struct{}{}
	}

	return portBindings, exposedPorts, nil
}

// CommitToImage 将容器提交为镜像
func (cs *ContainerService) CommitToImage(containerName string, imageName string) map[string]interface{} {
	ctx := context.Background()
	commit, err := cs.client.ContainerCommit(ctx, containerName, container.CommitOptions{Reference: imageName})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success", "id": commit.ID}
}
