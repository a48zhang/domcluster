package services

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// DockerHandler Docker 命令处理器
type DockerHandler struct {
	server *DomclusterServer
}

// NewDockerHandler 创建 Docker 处理器
func NewDockerHandler(server *DomclusterServer) *DockerHandler {
	return &DockerHandler{
		server: server,
	}
}

// executeDockerCommand 执行 Docker 命令的通用方法
func (h *DockerHandler) executeDockerCommand(ctx context.Context, nodeID, command string, data map[string]interface{}) ([]byte, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	// 发送请求
	go func() {
		dataBytes, err := json.Marshal(data)
		if err != nil {
			select {
			case errChan <- fmt.Errorf("failed to marshal data: %w", err):
			case <-ctx.Done():
			}
			return
		}

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, command, reqID, dataBytes, resultChan, errChan); err != nil {
			select {
			case errChan <- err:
			case <-ctx.Done():
			}
			return
		}
	}()

	// 等待结果或超时
	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		return result.Data, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// ListContainers 列出指定节点的容器
func (h *DockerHandler) ListContainers(ctx context.Context, nodeID string, all bool) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"all": all,
	}
	resultData, err := h.executeDockerCommand(ctx, nodeID, "docker_list", data)
	if err != nil {
		return nil, err
	}

	var containers []map[string]interface{}
	if err := json.Unmarshal(resultData, &containers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal containers: %w", err)
	}
	return map[string]interface{}{
		"containers": containers,
	}, nil
}

// StartContainer 启动容器
func (h *DockerHandler) StartContainer(ctx context.Context, nodeID, containerID string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"container_id": containerID,
	}
	resultData, err := h.executeDockerCommand(ctx, nodeID, "docker_start", data)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(resultData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return resp, nil
}

// StopContainer 停止容器
func (h *DockerHandler) StopContainer(ctx context.Context, nodeID, containerID string, timeout int) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"container_id": containerID,
		"timeout":      timeout,
	}
	resultData, err := h.executeDockerCommand(ctx, nodeID, "docker_stop", data)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(resultData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return resp, nil
}

// RestartContainer 重启容器
func (h *DockerHandler) RestartContainer(ctx context.Context, nodeID, containerID string, timeout int) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"container_id": containerID,
		"timeout":      timeout,
	}
	resultData, err := h.executeDockerCommand(ctx, nodeID, "docker_restart", data)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(resultData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return resp, nil
}

// GetContainerLogs 获取容器日志
func (h *DockerHandler) GetContainerLogs(ctx context.Context, nodeID, containerID, tail string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"container_id": containerID,
		"tail":         tail,
	}
	resultData, err := h.executeDockerCommand(ctx, nodeID, "docker_logs", data)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(resultData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return resp, nil
}

// GetContainerStats 获取容器统计信息
func (h *DockerHandler) GetContainerStats(ctx context.Context, nodeID, containerID string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"container_id": containerID,
	}
	resultData, err := h.executeDockerCommand(ctx, nodeID, "docker_stats", data)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(resultData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return resp, nil
}

// InspectContainer 查看容器详情
func (h *DockerHandler) InspectContainer(ctx context.Context, nodeID, containerID string) (map[string]interface{}, error) {
	data := map[string]interface{}{
		"container_id": containerID,
	}
	resultData, err := h.executeDockerCommand(ctx, nodeID, "docker_inspect", data)
	if err != nil {
		return nil, err
	}

	var resp map[string]interface{}
	if err := json.Unmarshal(resultData, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}
	return resp, nil
}

// sendDockerCommand 发送 Docker 命令到指定节点
func (h *DockerHandler) sendDockerCommand(nodeID, cmd, reqID string, data []byte, resultChan chan *DockerResult, errChan chan error) error {
	// 注册响应处理器
	h.server.RegisterDockerResponse(reqID, resultChan)

	// 发送命令到节点
	if err := h.server.SendToNode(nodeID, cmd, reqID, data); err != nil {
		errChan <- fmt.Errorf("failed to send command to node: %w", err)
		return err
	}

	zap.L().Sugar().Infof("Sent docker command %s to node %s", cmd, nodeID)
	return nil
}

// DockerResult Docker 命令结果
type DockerResult struct {
	Status int32
	Data   []byte
}

// generateReqID 生成请求 ID
func generateReqID() string {
	return fmt.Sprintf("docker_%d", time.Now().UnixNano())
}