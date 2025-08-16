package service

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/network"
)

// NetworkService 网络服务
type NetworkService struct {
	*DockerService
}

// NewNetworkService 创建网络服务
func NewNetworkService(dockerService *DockerService) *NetworkService {
	return &NetworkService{DockerService: dockerService}
}

// Create 创建网络
func (ns *NetworkService) Create(networkName string, isInternal bool, ipamConfig []network.IPAMConfig) map[string]interface{} {
	ctx := context.Background()
	_, err := ns.client.NetworkCreate(ctx, networkName, types.NetworkCreate{
		Driver:   "bridge",
		IPAM:     &network.IPAM{Config: ipamConfig},
		Internal: isInternal,
	})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// Delete 删除网络
func (ns *NetworkService) Delete(networkName string) map[string]interface{} {
	ctx := context.Background()
	err := ns.client.NetworkRemove(ctx, networkName)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// List 列出所有网络
func (ns *NetworkService) List() ([]types.NetworkResource, error) {
	ctx := context.Background()
	networks, err := ns.client.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	return networks, nil
}

// GetInfo 获取网络信息
func (ns *NetworkService) GetInfo(networkName string) (types.NetworkResource, error) {
	ctx := context.Background()
	network, err := ns.client.NetworkInspect(ctx, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return types.NetworkResource{}, err
	}
	return network, nil
}

// Connect 连接容器到网络
func (ns *NetworkService) Connect(networkName string, containerName string, ipv4Address string) map[string]interface{} {
	ctx := context.Background()
	err := ns.client.NetworkConnect(ctx, networkName, containerName, &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: ipv4Address,
		},
	})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

// Disconnect 断开容器与网络的连接
func (ns *NetworkService) Disconnect(networkName string, containerName string) map[string]interface{} {
	ctx := context.Background()
	err := ns.client.NetworkDisconnect(ctx, networkName, containerName, true)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}
