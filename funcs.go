package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/system"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/lithammer/fuzzysearch/fuzzy"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	rawNet "net"
)

// cleanANSIEscapes 清理 ANSI 转义序列
func cleanANSIEscapes(text string) string {
	// 匹配所有 ANSI 转义序列的正则表达式
	ansiRegex := regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)
	// 移除所有 ANSI 转义序列
	cleaned := ansiRegex.ReplaceAllString(text, "")
	return cleaned
}

func GetDockerClient(remote string) *client.Client {
	if remote == "" {
		cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		return cli
	} else {
		cli, err := client.NewClientWithOpts(client.WithHost(remote), client.WithAPIVersionNegotiation())
		if err != nil {
			panic(err)
		}
		return cli
	}
}

func DeleteImage(cli *client.Client, imageName string) map[string]interface{} {
	ctx := context.Background()
	_, err := cli.ImageRemove(ctx, imageName, types.ImageRemoveOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func GetImageInfo(cli *client.Client, imageName string) (types.ImageInspect, error) {
	ctx := context.Background()
	info, _, err := cli.ImageInspectWithRaw(ctx, imageName)
	if err != nil {
		return types.ImageInspect{}, err
	}
	return info, nil
}

func PullImage(cli *client.Client, imageName string) map[string]interface{} {
	ctx := context.Background()
	out, err := cli.ImagePull(ctx, imageName, types.ImagePullOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	defer out.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	return map[string]interface{}{"status": "success", "message": buf.String()}
}

func ListImages(cli *client.Client) ([]image.Summary, error) {
	ctx := context.Background()
	images, err := cli.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		return nil, err
	}
	return images, nil
}

func SearchImages(cli *client.Client, imageName string) []image.Summary {
	allImages, _ := ListImages(cli)
	var images []image.Summary
	for _, img := range allImages {
		// 模糊搜索
		if fuzzy.Match(imageName, img.RepoTags[0]) {
			images = append(images, img)
		}
	}
	return images
}

func RestartContainer(cli *client.Client, containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cli.ContainerRestart(ctx, containerName, container.StopOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func StartContainer(cli *client.Client, containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cli.ContainerStart(ctx, containerName, container.StartOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func StopContainer(cli *client.Client, containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cli.ContainerStop(ctx, containerName, container.StopOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func GetDockerInfo(cli *client.Client) (system.Info, error) {
	ctx := context.Background()
	info, err := cli.Info(ctx)
	if err != nil {
		return system.Info{}, err
	}
	return info, nil
}

func GetDockerVersion(cli *client.Client) (types.Version, error) {
	ctx := context.Background()
	version, err := cli.ServerVersion(ctx)
	if err != nil {
		return types.Version{}, err
	}
	return version, nil
}

func GetRunningContainers(cli *client.Client) ([]types.Container, error) {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, container.ListOptions{})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func GetAllContainers(cli *client.Client) ([]types.Container, error) {
	ctx := context.Background()
	containers, err := cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return nil, err
	}
	return containers, nil
}

func SearchContainers(cli *client.Client, containerName string) []types.Container {
	allContainers, _ := GetAllContainers(cli)
	var containers []types.Container
	for _, container := range allContainers {
		// 模糊搜索
		if fuzzy.Match(containerName, container.Names[0]) {
			containers = append(containers, container)
		}
	}
	return containers
}

func GetContainerInfo(cli *client.Client, containerName string) (types.ContainerJSON, error) {
	ctx := context.Background()
	c, err := cli.ContainerInspect(ctx, containerName)
	if err != nil {
		return types.ContainerJSON{}, err
	}
	return c, nil
}

// 新运行一个容器
// imageName: "ubuntu:latest"
// containerName: "container_123"
// cmd: "/bin/bash"
// addCaps: []string{"SYS_ADMIN"}
// hostname: "container_123"
// portMap: []string{"10080:80", "18080:8080"} 冒号前是宿主机端口，冒号后是容器端口
func RunContainer(cli *client.Client, imageName string, containerName string, cmd string, addCaps []string, hostname string, portMap []string) (types.ContainerJSON, map[string]interface{}) {
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

	// 如果DefaultNet不存在则创建
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "list networks error", "raw_message": err.Error()}
	}
	defaultNetExist := false
	for _, network := range networks {
		if network.Name == "DefaultNet" {
			defaultNetExist = true
			break
		}
	}
	if !defaultNetExist {
		_, err := cli.NetworkCreate(ctx, "DefaultNet", types.NetworkCreate{})
		if err != nil {
			return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "create DefaultNet error", "raw_message": err.Error()}
		}
	}
	// 检测hostname
	if hostname == "" {
		hostname = containerName
	}
	// 检查端口映射
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for _, p := range portMap {
		if strings.Count(p, ":") != 1 {
			return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "portMap error"}
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
	response, err := cli.ContainerCreate(ctx, &container.Config{
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
	StartContainer(cli, response.ID)
	t, err := cli.ContainerInspect(ctx, response.ID)
	if err != nil {
		return types.ContainerJSON{}, map[string]interface{}{"status": "fail", "message": "inspect container error", "raw_message": err.Error()}
	}
	return t, map[string]interface{}{"status": "success"}
}

func DeleteContainer(cli *client.Client, containerName string) map[string]interface{} {
	ctx := context.Background()
	// 删除容器
	err := cli.ContainerRemove(ctx, containerName, container.RemoveOptions{})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": "Remove Container Error", "raw_message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func GetContainerLogs(cli *client.Client, containerName string) (string, error) {
	ctx := context.Background()
	options := container.LogsOptions{ShowStdout: true, ShowStderr: true}
	out, err := cli.ContainerLogs(ctx, containerName, options)
	if err != nil {
		return "", err
	}
	defer out.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(out)
	return buf.String(), nil
}

func ContainerToImage(cli *client.Client, containerName string, imageName string) map[string]interface{} {
	ctx := context.Background()
	commit, err := cli.ContainerCommit(ctx, containerName, container.CommitOptions{Reference: imageName})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success", "id": commit.ID}
}

func GetDockerEvents(cli *client.Client) (<-chan events.Message, <-chan error) {
	ctx := context.Background()
	options := types.EventsOptions{}
	messages, errs := cli.Events(ctx, options)
	return messages, errs
}

func UpdateContainer(cli *client.Client, containerName string, memery int64, cpu int64) map[string]interface{} {
	ctx := context.Background()
	updateConfig := container.UpdateConfig{
		Resources: container.Resources{
			Memory:     memery,
			MemorySwap: memery * 2,
			CPUShares:  cpu,
		},
	}
	_, err := cli.ContainerUpdate(ctx, containerName, updateConfig)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}

}

func CopyFromContainer(containerName string, srcPath string, dstPath string) map[string]interface{} {
	// 调用docker cp
	cmd := exec.Command("docker", "cp", containerName+":"+srcPath, dstPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": stderr.String()}
	}
	return map[string]interface{}{"status": "success"}
}

func CopyToContainer(containerName string, srcPath string, dstPath string) map[string]interface{} {
	// 调用docker cp
	cmd := exec.Command("docker", "cp", srcPath, containerName+":"+dstPath)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": stderr.String()}
	}
	return map[string]interface{}{"status": "success"}
}

func CreateNetwork(cli *client.Client, networkName string, isInternal bool, IPAMConfig []network.IPAMConfig) map[string]interface{} {
	ctx := context.Background()
	_, err := cli.NetworkCreate(ctx, networkName, types.NetworkCreate{
		Driver:   "bridge",
		IPAM:     &network.IPAM{Config: IPAMConfig},
		Internal: isInternal,
	})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func DeleteNetwork(cli *client.Client, networkName string) map[string]interface{} {
	ctx := context.Background()
	err := cli.NetworkRemove(ctx, networkName)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func ListNetworks(cli *client.Client) ([]types.NetworkResource, error) {
	ctx := context.Background()
	networks, err := cli.NetworkList(ctx, types.NetworkListOptions{})
	if err != nil {
		return nil, err
	}
	return networks, nil
}

func NetworkInfo(cli *client.Client, networkName string) (types.NetworkResource, error) {
	ctx := context.Background()
	network, err := cli.NetworkInspect(ctx, networkName, types.NetworkInspectOptions{})
	if err != nil {
		return types.NetworkResource{}, err
	}
	return network, nil
}

func ConnectNetwork(cli *client.Client, networkName string, containerName string, ipv4Address string) map[string]interface{} {
	ctx := context.Background()
	err := cli.NetworkConnect(ctx, networkName, containerName, &network.EndpointSettings{
		IPAMConfig: &network.EndpointIPAMConfig{
			IPv4Address: ipv4Address,
		},
	})
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func DisconnectNetwork(cli *client.Client, networkName string, containerName string) map[string]interface{} {
	ctx := context.Background()
	err := cli.NetworkDisconnect(ctx, networkName, containerName, true)
	if err != nil {
		return map[string]interface{}{"status": "fail", "message": err.Error()}
	}
	return map[string]interface{}{"status": "success"}
}

func TrivyInstall() map[string]interface{} {
	// 安装trivy
	// 识别系统
	if runtime.GOOS != "linux" {
		return map[string]interface{}{"status": "fail", "code": -1, "message": "only support linux"}
	}
	// source /etc/os-release && echo $ID
	cmd := exec.Command("bash", "-c", "source /etc/os-release && echo $ID")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]interface{}{"status": "fail", "code": -2, "message": "get os id fail", "raw_message": stderr.String()}
	}
	osID := stdout.String()[0 : len(stdout.String())-1]
	// 安装
	if osID == "ubuntu" || osID == "debian" || osID == "kylin" {
		cmd = exec.Command("bash", "-c", "wget https://alist.jz-home.top/d/%E9%98%BF%E9%87%8C%E4%BA%91%E7%9B%98/trivydb/trivy_0.49.1_Linux-64bit.deb")
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return map[string]interface{}{"status": "fail", "code": -3, "message": "download trivy deb fail", "raw_message": stderr.String()}
		}
		cmd = exec.Command("bash", "-c", "dpkg -i trivy_0.49.1_Linux-64bit.deb")
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			return map[string]interface{}{"status": "fail", "code": -4, "message": "install trivy deb fail", "raw_message": stderr.String()}
		}
	} else {
		if osID == "centos" || osID == "rhel" {
			cmd = exec.Command("bash", "-c", "wget https://alist.jz-home.top/d/%E9%98%BF%E9%87%8C%E4%BA%91%E7%9B%98/trivydb/trivy_0.49.1_Linux-64bit.rpm")
			var stdout, stderr bytes.Buffer
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				return map[string]interface{}{"status": "fail", "code": -5, "message": "download trivy rpm fail", "raw_message": stderr.String()}
			}
			cmd = exec.Command("bash", "-c", "rpm -ivh trivy_0.49.1_Linux-64bit.rpm")
			cmd.Stdout = &stdout
			cmd.Stderr = &stderr
			err = cmd.Run()
			if err != nil {
				return map[string]interface{}{"status": "fail", "code": -6, "message": "install trivy rpm fail", "raw_message": stderr.String()}
			}
		} else {
			return map[string]interface{}{"status": "fail", "code": -7, "message": "OS " + osID + " is not supported"}
		}

	}
	return map[string]interface{}{"status": "success", "code": 0, "message": "success"}
}

func DownloadTrivyDB() map[string]interface{} {
	// 下载trivy漏洞库
	cmd := exec.Command("wget", "https://alist.jz-home.top/d/%E9%98%BF%E9%87%8C%E4%BA%91%E7%9B%98/trivydb/db.tar.gz")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]interface{}{"status": "fail", "code": -1, "message": "wget db failed", "raw_message": stderr.String()}
	}
	// 解压
	os.MkdirAll("/root/.cache/trivy/db", os.ModePerm)
	cmd = exec.Command("tar", "-zxf", "db.tar.gz", "-C", "/root/.cache/trivy/db")
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err = cmd.Run()
	if err != nil {
		return map[string]interface{}{"status": "fail", "code": -2, "message": "tar -zxf failed", "raw_message": stderr.String()}
	}
	err = os.Remove("db.tar.gz")
	if err != nil {
		return map[string]interface{}{"status": "fail", "code": -3, "message": "remove db.tar.gz failed", "raw_message": err.Error()}
	}
	return map[string]interface{}{"status": "success", "code": 0, "message": "success"}
}

func RemoveTrivyDB() map[string]interface{} {
	// 删除trivy漏洞库
	cmd := exec.Command("rm", "-rf", "/root/.cache/trivy/db")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]interface{}{"status": "fail", "code": -1, "message": stderr.String()}
	}
	return map[string]interface{}{"status": "success", "code": 0, "message": "success"}
}

func GetTrivyDBInfo() map[string]interface{} {
	// 获取trivy漏洞库信息
	cmd := exec.Command("trivy", "-v")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]interface{}{"status": "fail", "code": -1, "message": stderr.String()}
	}
	return map[string]interface{}{"status": "success", "code": 0, "message": "success", "data": stdout.String()}
}

func ScanImage(imageName string) (map[string]interface{}, string) {
	// 调用trivy扫描镜像漏洞
	info := GetTrivyDBInfo()
	if info["status"] == "fail" {
		return map[string]interface{}{}, "trivy db not exist"
	}
	cmd := exec.Command("trivy", "image", "--skip-db-update", "--offline-scan", "-f", "json", "--scanners", "vuln", imageName)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return map[string]interface{}{}, stderr.String()
	}
	r := make(map[string]interface{})
	if json.Unmarshal(stdout.Bytes(), &r) != nil {
		return map[string]interface{}{}, "json unmarshal error"
	}
	return r, ""
}

func DockerExec(cli *client.Client, containerName string, cmd string) (map[string]interface{}, error) {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/sh"},
		Tty:          true,
	}

	resp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return map[string]interface{}{"status": "fail", "error": err.Error()}, err
	}

	hijack, err := cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{
		Detach: false,
		Tty:    true,
	})
	if err != nil {
		return map[string]interface{}{"status": "fail", "error": err.Error()}, err
	}
	defer hijack.Close()

	hijack.Conn.Write([]byte(cmd + "\nexit\n"))

	scanner := bufio.NewScanner(hijack.Reader)
	output := ""
	for scanner.Scan() {
		output += scanner.Text() + "\n"
	}

	return map[string]interface{}{"status": "success", "data": output}, nil
}

// DockerExecInteractive 创建一个交互式TTY会话，返回一个可以持续执行命令的会话对象
type ExecSession struct {
	ContainerName string
	ExecID        string
	Hijack        types.HijackedResponse
	Cli           *client.Client
	Ctx           context.Context
}

// CreateExecSession 创建一个新的交互式执行会话
func CreateExecSession(cli *client.Client, containerName string) (*ExecSession, error) {
	ctx := context.Background()

	execConfig := types.ExecConfig{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Cmd:          []string{"/bin/bash"},
		Tty:          true,
	}

	resp, err := cli.ContainerExecCreate(ctx, containerName, execConfig)
	if err != nil {
		return nil, err
	}

	hijack, err := cli.ContainerExecAttach(ctx, resp.ID, types.ExecStartCheck{
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
		Cli:           cli,
		Ctx:           ctx,
	}

	return session, nil
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
	cleanedOutput := cleanANSIEscapes(output.String())
	return cleanedOutput, nil
}

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

    cleanedOutput := cleanANSIEscapes(output.String())
    return cleanedOutput, nil
}

// Close 关闭会话
func (s *ExecSession) Close() {
	// 发送exit命令
	s.Hijack.Conn.Write([]byte("exit\n"))
	// 关闭连接
	s.Hijack.Close()
}

func GetProcess(containerName string) (map[string]interface{}, error) {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	// 获取容器内进程
	ctx := context.Background()
	s, err := cli.ContainerTop(ctx, containerName, []string{"aux"})
	if err != nil {
		return nil, err
	}
	r := make(map[string]interface{})
	for _, v := range s.Titles {
		r[v] = make([]string, 0)
	}
	for _, v := range s.Processes {
		for i, vv := range v {
			r[s.Titles[i]] = append(r[s.Titles[i]].([]string), vv)
		}
	}
	return r, nil
}

func GetContainerStats(cli *client.Client, containerName string) (map[string]interface{}, string) {
	ctx := context.Background()
	// 获取容器状态
	stats, err := cli.ContainerStats(ctx, containerName, false)
	if err != nil {
		return nil, err.Error()
	}
	defer stats.Body.Close()
	buf := new(bytes.Buffer)
	buf.ReadFrom(stats.Body)
	r := make(map[string]interface{})
	if json.Unmarshal(buf.Bytes(), &r) != nil {
		return nil, "json unmarshal error"
	}
	return r, ""
}

func GetHostCpuPercent() float64 {
	percent, _ := cpu.Percent(time.Second, false)
	return percent[0]
}

func GetHostMemeryPercent() float64 {
	v, _ := mem.VirtualMemory()
	return v.UsedPercent
}

func GetHostDiskPercent() float64 {
	v, _ := disk.Usage("/")
	return v.UsedPercent
}

func GetHostNetSpeed() (float64, float64) {
	// 获取网络速度，单位为字节
	// 获取上行流量
	up, _ := net.IOCounters(true)
	upSpeed := up[0].BytesSent
	// 获取下行流量
	down, _ := net.IOCounters(true)
	downSpeed := down[0].BytesRecv
	// 等待1s
	time.Sleep(time.Second)
	// 获取上行流量
	up, _ = net.IOCounters(true)
	upSpeed = up[0].BytesSent - upSpeed
	// 获取下行流量
	down, _ = net.IOCounters(true)
	downSpeed = down[0].BytesRecv - downSpeed
	return float64(upSpeed), float64(downSpeed)
}

func RunContainerByCompose(composeFile string, otherFiles map[string]string, deleteTempDir bool) map[string]interface{} {
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
