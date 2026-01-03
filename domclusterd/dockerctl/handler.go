package dockerctl

import (
	"encoding/json"
	"fmt"
)

// Handler Docker 操作处理器
type Handler struct {
	client *DockerClient
}

// NewHandler 创建 Docker 处理器
func NewHandler(client *DockerClient) *Handler {
	return &Handler{
		client: client,
	}
}

// ListContainers 列出容器
func (h *Handler) ListContainers(all bool) ([]byte, error) {
	containers, err := h.client.ListContainers(all)
	if err != nil {
		return nil, err
	}

	data, err := json.Marshal(containers)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal containers: %w", err)
	}

	return data, nil
}

// StartContainer 启动容器
func (h *Handler) StartContainer(containerID string) ([]byte, error) {
	err := h.client.StartContainer(containerID)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"message": "container started",
		"container_id": containerID,
	})
}

// StopContainer 停止容器
func (h *Handler) StopContainer(containerID string, timeout int) ([]byte, error) {
	err := h.client.StopContainer(containerID, timeout)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"message": "container stopped",
		"container_id": containerID,
	})
}

// RestartContainer 重启容器
func (h *Handler) RestartContainer(containerID string, timeout int) ([]byte, error) {
	err := h.client.RestartContainer(containerID, timeout)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"message": "container restarted",
		"container_id": containerID,
	})
}

// GetContainerLogs 获取容器日志
func (h *Handler) GetContainerLogs(containerID string, tail string) ([]byte, error) {
	logs, err := h.client.GetContainerLogs(containerID, tail)
	if err != nil {
		return nil, err
	}

	return json.Marshal(map[string]interface{}{
		"container_id": containerID,
		"logs": logs,
	})
}

// GetContainerStats 获取容器统计信息
func (h *Handler) GetContainerStats(containerID string) ([]byte, error) {
	return h.client.GetContainerStats(containerID)
}

// InspectContainer 查看容器详情
func (h *Handler) InspectContainer(containerID string) ([]byte, error) {
	container, err := h.client.InspectContainer(containerID)
	if err != nil {
		return nil, err
	}

	// 提取关键信息
	info := map[string]interface{}{
		"id": container.ID,
		"name": container.Name,
		"image": container.Config.Image,
		"state": container.State.Status,
		"created": container.Created,
		"restart_count": container.RestartCount,
		"ip": container.NetworkSettings.IPAddress,
	}

	return json.Marshal(info)
}

// HandleCommand 处理 Docker 命令
func (h *Handler) HandleCommand(cmd string, data map[string]interface{}) ([]byte, error) {
	switch cmd {
	case "docker_list":
		all := false
		if val, ok := data["all"].(bool); ok {
			all = val
		}
		return h.ListContainers(all)

	case "docker_start":
		containerID, ok := data["container_id"].(string)
		if !ok {
			return nil, fmt.Errorf("missing container_id")
		}
		return h.StartContainer(containerID)

	case "docker_stop":
		containerID, ok := data["container_id"].(string)
		if !ok {
			return nil, fmt.Errorf("missing container_id")
		}
		timeout := 10
		if val, ok := data["timeout"].(float64); ok {
			timeout = int(val)
		}
		return h.StopContainer(containerID, timeout)

	case "docker_restart":
		containerID, ok := data["container_id"].(string)
		if !ok {
			return nil, fmt.Errorf("missing container_id")
		}
		timeout := 10
		if val, ok := data["timeout"].(float64); ok {
			timeout = int(val)
		}
		return h.RestartContainer(containerID, timeout)

	case "docker_logs":
		containerID, ok := data["container_id"].(string)
		if !ok {
			return nil, fmt.Errorf("missing container_id")
		}
		tail := "100"
		if val, ok := data["tail"].(string); ok {
			tail = val
		}
		return h.GetContainerLogs(containerID, tail)

	case "docker_stats":
		containerID, ok := data["container_id"].(string)
		if !ok {
			return nil, fmt.Errorf("missing container_id")
		}
		return h.GetContainerStats(containerID)

	case "docker_inspect":
		containerID, ok := data["container_id"].(string)
		if !ok {
			return nil, fmt.Errorf("missing container_id")
		}
		return h.InspectContainer(containerID)

	default:
		return nil, fmt.Errorf("unknown docker command: %s", cmd)
	}
}