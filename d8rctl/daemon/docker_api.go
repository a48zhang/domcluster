package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"d8rctl/services"
)

const (
	// DefaultRequestTimeout 默认请求超时时间
	DefaultRequestTimeout = 60 * time.Second
)

// sendJSONError 发送 JSON 格式的错误响应
func sendJSONError(w http.ResponseWriter, statusCode int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// sendJSONResponse 发送 JSON 格式的成功响应
func sendJSONResponse(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(data)
}

// checkMethod 检查 HTTP 方法是否匹配
func checkMethod(w http.ResponseWriter, r *http.Request, method string) bool {
	if r.Method != method {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

// handleDockerList 处理列出容器请求
func (hs *HTTPServer) handleDockerList(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodGet) {
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		sendJSONError(w, http.StatusBadRequest, "node_id is required")
		return
	}

	allStr := r.URL.Query().Get("all")
	all := allStr == "true" || allStr == "1"

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(r.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.ListContainers(ctx, nodeID, all)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendJSONError(w, http.StatusGatewayTimeout, "request timeout")
		} else {
			sendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendJSONResponse(w, result)
}

// handleDockerStart 处理启动容器请求
func (hs *HTTPServer) handleDockerStart(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		sendJSONError(w, http.StatusBadRequest, "node_id and container_id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.StartContainer(ctx, req.NodeID, req.ContainerID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendJSONError(w, http.StatusGatewayTimeout, "request timeout")
		} else {
			sendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendJSONResponse(w, result)
}

// handleDockerStop 处理停止容器请求
func (hs *HTTPServer) handleDockerStop(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
		Timeout     int    `json:"timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		sendJSONError(w, http.StatusBadRequest, "node_id and container_id are required")
		return
	}

	if req.Timeout == 0 {
		req.Timeout = 10
	}

	ctx, cancel := context.WithTimeout(r.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.StopContainer(ctx, req.NodeID, req.ContainerID, req.Timeout)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendJSONError(w, http.StatusGatewayTimeout, "request timeout")
		} else {
			sendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendJSONResponse(w, result)
}

// handleDockerRestart 处理重启容器请求
func (hs *HTTPServer) handleDockerRestart(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodPost) {
		return
	}

	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
		Timeout     int    `json:"timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "invalid request")
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		sendJSONError(w, http.StatusBadRequest, "node_id and container_id are required")
		return
	}

	if req.Timeout == 0 {
		req.Timeout = 10
	}

	ctx, cancel := context.WithTimeout(r.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.RestartContainer(ctx, req.NodeID, req.ContainerID, req.Timeout)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendJSONError(w, http.StatusGatewayTimeout, "request timeout")
		} else {
			sendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendJSONResponse(w, result)
}

// handleDockerLogs 处理获取容器日志请求
func (hs *HTTPServer) handleDockerLogs(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodGet) {
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	containerID := r.URL.Query().Get("container_id")
	tail := r.URL.Query().Get("tail")

	if nodeID == "" || containerID == "" {
		sendJSONError(w, http.StatusBadRequest, "node_id and container_id are required")
		return
	}

	if tail == "" {
		tail = "100"
	}

	ctx, cancel := context.WithTimeout(r.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.GetContainerLogs(ctx, nodeID, containerID, tail)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendJSONError(w, http.StatusGatewayTimeout, "request timeout")
		} else {
			sendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendJSONResponse(w, result)
}

// handleDockerStats 处理获取容器统计信息请求
func (hs *HTTPServer) handleDockerStats(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodGet) {
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	containerID := r.URL.Query().Get("container_id")

	if nodeID == "" || containerID == "" {
		sendJSONError(w, http.StatusBadRequest, "node_id and container_id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.GetContainerStats(ctx, nodeID, containerID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendJSONError(w, http.StatusGatewayTimeout, "request timeout")
		} else {
			sendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendJSONResponse(w, result)
}

// handleDockerInspect 处理查看容器详情请求
func (hs *HTTPServer) handleDockerInspect(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodGet) {
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	containerID := r.URL.Query().Get("container_id")

	if nodeID == "" || containerID == "" {
		sendJSONError(w, http.StatusBadRequest, "node_id and container_id are required")
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.InspectContainer(ctx, nodeID, containerID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			sendJSONError(w, http.StatusGatewayTimeout, "request timeout")
		} else {
			sendJSONError(w, http.StatusInternalServerError, err.Error())
		}
		return
	}

	sendJSONResponse(w, result)
}

// handleDockerNodes 处理获取所有节点列表
func (hs *HTTPServer) handleDockerNodes(w http.ResponseWriter, r *http.Request) {
	if !checkMethod(w, r, http.MethodGet) {
		return
	}

	nodes := hs.svc.(*services.DomclusterServer).GetNodeManager().ListNodes()
	sendJSONResponse(w, map[string]interface{}{
		"nodes": nodes,
	})
}