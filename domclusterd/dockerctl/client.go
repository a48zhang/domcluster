package dockerctl

import (
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"go.uber.org/zap"
)

// ContainerInfo 容器信息
type ContainerInfo struct {
	ID         string
	Name       string
	Image      string
	Status     string
	State      string
	Created    int64
	Ports      []types.Port
	Networks   []string
	Mounts     []types.MountPoint
}

// DockerClient Docker 客户端封装
type DockerClient struct {
	cli           *client.Client
	defaultTimeout time.Duration
}

// NewDockerClient 创建 Docker 客户端
func NewDockerClient() (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	_, err = cli.Ping(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to docker daemon: %w", err)
	}

	zap.L().Sugar().Info("Successfully connected to Docker daemon")
	return &DockerClient{
		cli:           cli,
		defaultTimeout: 10 * time.Second,
	}, nil
}

// ListContainers 列出所有容器
func (dc *DockerClient) ListContainers(all bool) ([]ContainerInfo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dc.defaultTimeout)
	defer cancel()
	
	containers, err := dc.cli.ContainerList(ctx, container.ListOptions{All: all})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []ContainerInfo
	for _, c := range containers {
		var networks []string
		for networkName := range c.NetworkSettings.Networks {
			networks = append(networks, networkName)
		}

		result = append(result, ContainerInfo{
			ID:       c.ID[:12],
			Name:     c.Names[0],
			Image:    c.Image,
			Status:   c.Status,
			State:    c.State,
			Created:  c.Created,
			Ports:    c.Ports,
			Networks: networks,
			Mounts:   c.Mounts,
		})
	}

	return result, nil
}

// InspectContainer 查看容器详情
func (dc *DockerClient) InspectContainer(containerID string) (types.ContainerJSON, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dc.defaultTimeout)
	defer cancel()
	
	container, err := dc.cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return types.ContainerJSON{}, fmt.Errorf("failed to inspect container: %w", err)
	}
	return container, nil
}

// StartContainer 启动容器
func (dc *DockerClient) StartContainer(containerID string) error {
	ctx, cancel := context.WithTimeout(context.Background(), dc.defaultTimeout)
	defer cancel()
	
	err := dc.cli.ContainerStart(ctx, containerID, container.StartOptions{})
	if err != nil {
		return fmt.Errorf("failed to start container: %w", err)
	}
	zap.L().Sugar().Infof("Container %s started", containerID)
	return nil
}

// StopContainer 停止容器
func (dc *DockerClient) StopContainer(containerID string, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dc.defaultTimeout)
	defer cancel()
	
	err := dc.cli.ContainerStop(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to stop container: %w", err)
	}
	zap.L().Sugar().Infof("Container %s stopped", containerID)
	return nil
}

// RestartContainer 重启容器
func (dc *DockerClient) RestartContainer(containerID string, timeout int) error {
	ctx, cancel := context.WithTimeout(context.Background(), dc.defaultTimeout)
	defer cancel()
	
	err := dc.cli.ContainerRestart(ctx, containerID, container.StopOptions{
		Timeout: &timeout,
	})
	if err != nil {
		return fmt.Errorf("failed to restart container: %w", err)
	}
	zap.L().Sugar().Infof("Container %s restarted", containerID)
	return nil
}

// RemoveContainer 删除容器
func (dc *DockerClient) RemoveContainer(containerID string, force bool) error {
	ctx := context.Background()
	err := dc.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{
		Force: force,
	})
	if err != nil {
		return fmt.Errorf("failed to remove container: %w", err)
	}
	zap.L().Sugar().Infof("Container %s removed", containerID)
	return nil
}

// GetContainerLogs 获取容器日志
func (dc *DockerClient) GetContainerLogs(containerID string, tail string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dc.defaultTimeout)
	defer cancel()
	
	options := container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Tail:       tail,
	}

	reader, err := dc.cli.ContainerLogs(ctx, containerID, options)
	if err != nil {
		return "", fmt.Errorf("failed to get container logs: %w", err)
	}
	defer reader.Close()

	logs := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			break
		}
		logs = append(logs, buf[:n]...)
	}

	return string(logs), nil
}

// GetContainerStats 获取容器统计信息
func (dc *DockerClient) GetContainerStats(containerID string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), dc.defaultTimeout)
	defer cancel()
	
	stats, err := dc.cli.ContainerStats(ctx, containerID, false)
	if err != nil {
		return nil, fmt.Errorf("failed to get container stats: %w", err)
	}
	defer stats.Body.Close()

	// 读取原始 JSON
	result := make([]byte, 0)
	buf := make([]byte, 1024)
	for {
		n, err := stats.Body.Read(buf)
		if err != nil {
			break
		}
		result = append(result, buf[:n]...)
	}

	return result, nil
}

// Close 关闭客户端连接
func (dc *DockerClient) Close() error {
	return dc.cli.Close()
}