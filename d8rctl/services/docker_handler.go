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

// ListContainers 列出指定节点的容器
func (h *DockerHandler) ListContainers(ctx context.Context, nodeID string, all bool) (map[string]interface{}, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	// 发送请求
	go func() {
		data := map[string]interface{}{
			"all": all,
		}
		dataBytes, _ := json.Marshal(data)

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, "docker_list", reqID, dataBytes, resultChan, errChan); err != nil {
			errChan <- err
		}
	}()

	// 等待结果或超时
	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		var containers []map[string]interface{}
		if err := json.Unmarshal(result.Data, &containers); err != nil {
			return nil, fmt.Errorf("failed to unmarshal containers: %w", err)
		}
		return map[string]interface{}{
			"containers": containers,
		}, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// StartContainer 启动容器
func (h *DockerHandler) StartContainer(ctx context.Context, nodeID, containerID string) (map[string]interface{}, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	go func() {
		data := map[string]interface{}{
			"container_id": containerID,
		}
		dataBytes, _ := json.Marshal(data)

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, "docker_start", reqID, dataBytes, resultChan, errChan); err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(result.Data, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// StopContainer 停止容器
func (h *DockerHandler) StopContainer(ctx context.Context, nodeID, containerID string, timeout int) (map[string]interface{}, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	go func() {
		data := map[string]interface{}{
			"container_id": containerID,
			"timeout":      timeout,
		}
		dataBytes, _ := json.Marshal(data)

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, "docker_stop", reqID, dataBytes, resultChan, errChan); err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(result.Data, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// RestartContainer 重启容器
func (h *DockerHandler) RestartContainer(ctx context.Context, nodeID, containerID string, timeout int) (map[string]interface{}, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	go func() {
		data := map[string]interface{}{
			"container_id": containerID,
			"timeout":      timeout,
		}
		dataBytes, _ := json.Marshal(data)

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, "docker_restart", reqID, dataBytes, resultChan, errChan); err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(result.Data, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetContainerLogs 获取容器日志
func (h *DockerHandler) GetContainerLogs(ctx context.Context, nodeID, containerID, tail string) (map[string]interface{}, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	go func() {
		data := map[string]interface{}{
			"container_id": containerID,
			"tail":         tail,
		}
		dataBytes, _ := json.Marshal(data)

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, "docker_logs", reqID, dataBytes, resultChan, errChan); err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(result.Data, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// GetContainerStats 获取容器统计信息
func (h *DockerHandler) GetContainerStats(ctx context.Context, nodeID, containerID string) (map[string]interface{}, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	go func() {
		data := map[string]interface{}{
			"container_id": containerID,
		}
		dataBytes, _ := json.Marshal(data)

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, "docker_stats", reqID, dataBytes, resultChan, errChan); err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(result.Data, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// InspectContainer 查看容器详情
func (h *DockerHandler) InspectContainer(ctx context.Context, nodeID, containerID string) (map[string]interface{}, error) {
	resultChan := make(chan *DockerResult, 1)
	errChan := make(chan error, 1)

	go func() {
		data := map[string]interface{}{
			"container_id": containerID,
		}
		dataBytes, _ := json.Marshal(data)

		reqID := generateReqID()
		if err := h.sendDockerCommand(nodeID, "docker_inspect", reqID, dataBytes, resultChan, errChan); err != nil {
			errChan <- err
		}
	}()

	select {
	case result := <-resultChan:
		if result.Status != 0 {
			return nil, fmt.Errorf("docker command failed: %s", string(result.Data))
		}
		var resp map[string]interface{}
		if err := json.Unmarshal(result.Data, &resp); err != nil {
			return nil, fmt.Errorf("failed to unmarshal response: %w", err)
		}
		return resp, nil
	case err := <-errChan:
		return nil, err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
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