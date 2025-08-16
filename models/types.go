package models

// StandardResponse 标准响应格式
type StandardResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

// ContainerRequest 创建容器请求
type ContainerRequest struct {
	ImageName     string   `json:"image_name"`
	ContainerName string   `json:"container_name"`
	Cmd           string   `json:"cmd"`
	AddCaps       []string `json:"add_caps"`
	HostName      string   `json:"host_name"`
	PortMap       []string `json:"port_map"`
}

// ContainerUpdateRequest 更新容器资源请求
type ContainerUpdateRequest struct {
	ContainerName string `json:"container_name"`
	Memory        int64  `json:"memory"`
	CPUShares     int64  `json:"cpu_shares"`
}

// NetworkRequest 网络创建请求
type NetworkRequest struct {
	NetworkName string `json:"network_name"`
	IsInternal  bool   `json:"is_internal"`
	Subnet      string `json:"subnet"`
	IPRange     string `json:"ip_range"`
	Gateway     string `json:"gateway"`
}

// NetworkConnectRequest 网络连接请求
type NetworkConnectRequest struct {
	NetworkName   string `json:"network_name"`
	ContainerName string `json:"container_name"`
	IPv4Address   string `json:"ipv4_address"`
}

// ExecRequest 命令执行请求
type ExecRequest struct {
	ContainerName string `json:"container_name"`
	Cmd           string `json:"cmd"`
	ExecSessionID string `json:"exec_session_id,omitempty"`
}

// FileOperationRequest 文件操作请求
type FileOperationRequest struct {
	ContainerName string `json:"container_name"`
	SourcePath    string `json:"source_path"`
	DestPath      string `json:"dest_path"`
}

// ComposeRequest Docker Compose请求
type ComposeRequest struct {
	ComposeFile   string            `json:"compose_file"`
	OtherFiles    map[string]string `json:"other_files"`
	DeleteTempDir bool              `json:"delete_temp_dir"`
}

// ImageRequest 镜像操作请求
type ImageRequest struct {
	ImageName string `json:"image_name"`
	Remote    string `json:"remote,omitempty"`
}

// ContainerNameRequest 通用容器名称请求
type ContainerNameRequest struct {
	ContainerName string `json:"container_name"`
}

// NetworkNameRequest 通用网络名称请求
type NetworkNameRequest struct {
	NetworkName string `json:"network_name"`
}

// SearchRequest 搜索请求
type SearchRequest struct {
	Query string `json:"query"`
}
