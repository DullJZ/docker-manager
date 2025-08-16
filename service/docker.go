package service

import (
	"context"

	"github.com/docker/docker/client"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/api/types"
)

// DockerService Docker 服务基础结构
type DockerService struct {
	client *client.Client
}

// NewDockerService 创建 Docker 服务
func NewDockerService(remote string) *DockerService {
	var cli *client.Client
	var err error

	if remote == "" {
		cli, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	} else {
		cli, err = client.NewClientWithOpts(client.WithHost(remote), client.WithAPIVersionNegotiation())
	}

	if err != nil {
		panic(err)
	}

	return &DockerService{
		client: cli,
	}
}

// GetClient 获取 Docker 客户端
func (ds *DockerService) GetClient() *client.Client {
	return ds.client
}

// GetDockerInfo 获取 Docker 信息
func (ds *DockerService) GetDockerInfo() (system.Info, error) {
	ctx := context.Background()
	info, err := ds.client.Info(ctx)
	if err != nil {
		return system.Info{}, err
	}
	return info, nil
}

func (ds *DockerService) GetDockerVersion() (types.Version, error) {
	ctx := context.Background()
	version, err := ds.client.ServerVersion(ctx)
	if err != nil {
		return types.Version{}, err
	}
	return version, nil
}
