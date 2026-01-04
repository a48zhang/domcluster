package daemon

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"d8rctl/services"
)

// handleDockerList 处理列出容器请求
func (hs *HTTPServer) handleDockerList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	if nodeID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "node_id is required"})
		return
	}

	allStr := r.URL.Query().Get("all")
	all := allStr == "true" || allStr == "1"

	// 创建带超时的 context
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.ListContainers(ctx, nodeID, all)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "request timeout"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerStart 处理启动容器请求
func (hs *HTTPServer) handleDockerStart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "node_id and container_id are required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.StartContainer(ctx, req.NodeID, req.ContainerID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "request timeout"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerStop 处理停止容器请求
func (hs *HTTPServer) handleDockerStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
		Timeout     int    `json:"timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "node_id and container_id are required"})
		return
	}

	if req.Timeout == 0 {
		req.Timeout = 10
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.StopContainer(ctx, req.NodeID, req.ContainerID, req.Timeout)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "request timeout"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerRestart 处理重启容器请求
func (hs *HTTPServer) handleDockerRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
		Timeout     int    `json:"timeout"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "invalid request"})
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "node_id and container_id are required"})
		return
	}

	if req.Timeout == 0 {
		req.Timeout = 10
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.RestartContainer(ctx, req.NodeID, req.ContainerID, req.Timeout)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "request timeout"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerLogs 处理获取容器日志请求
func (hs *HTTPServer) handleDockerLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	containerID := r.URL.Query().Get("container_id")
	tail := r.URL.Query().Get("tail")

	if nodeID == "" || containerID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "node_id and container_id are required"})
		return
	}

	if tail == "" {
		tail = "100"
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.GetContainerLogs(ctx, nodeID, containerID, tail)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "request timeout"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerStats 处理获取容器统计信息请求
func (hs *HTTPServer) handleDockerStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	containerID := r.URL.Query().Get("container_id")

	if nodeID == "" || containerID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "node_id and container_id are required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.GetContainerStats(ctx, nodeID, containerID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "request timeout"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerInspect 处理查看容器详情请求
func (hs *HTTPServer) handleDockerInspect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodeID := r.URL.Query().Get("node_id")
	containerID := r.URL.Query().Get("container_id")

	if nodeID == "" || containerID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{"error": "node_id and container_id are required"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 60*time.Second)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.InspectContainer(ctx, nodeID, containerID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		if ctx.Err() == context.DeadlineExceeded {
			w.WriteHeader(http.StatusGatewayTimeout)
			json.NewEncoder(w).Encode(map[string]string{"error": "request timeout"})
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// handleDockerNodes 处理获取所有节点列表
func (hs *HTTPServer) handleDockerNodes(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	nodes := hs.svc.(*services.DomclusterServer).GetNodeManager().ListNodes()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"nodes": nodes,
	})
}