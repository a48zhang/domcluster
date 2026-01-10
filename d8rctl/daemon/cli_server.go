package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"time"

	"d8rctl/services"
	"go.uber.org/zap"
)

const cliSocketPath = "/run/d8rctl/d8rctl.sock"

// CLIServer CLI 专用服务器（通过 Unix Domain Socket）
type CLIServer struct {
	server *http.Server
	svc    *services.DomclusterServer
}

// NewCLIServer 创建 CLI 服务器
func NewCLIServer(svc *services.DomclusterServer) *CLIServer {
	hs := &CLIServer{
		svc: svc,
	}

	mux := http.NewServeMux()

	// CLI 端点 - 不需要认证（通过 Unix Socket 访问，本身就是安全的）
	mux.HandleFunc("/status", hs.handleStatus)
	mux.HandleFunc("/stop", hs.handleStop)
	mux.HandleFunc("/restart", hs.handleRestart)
	mux.HandleFunc("/nodes", hs.handleNodes)
	mux.HandleFunc("/hosts/add", hs.handleAddHost)

	hs.server = &http.Server{
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	return hs
}

// Start 启动 CLI 服务器
func (cs *CLIServer) Start() error {
	// 删除已存在的 socket 文件
	if _, err := os.Stat(cliSocketPath); err == nil {
		os.Remove(cliSocketPath)
	}

	listener, err := net.Listen("unix", cliSocketPath)
	if err != nil {
		return fmt.Errorf("failed to listen on unix socket: %w", err)
	}

	// 设置 socket 文件权限为 0770，允许当前用户访问
	if err := os.Chmod(cliSocketPath, 0770); err != nil {
		zap.L().Sugar().Warnf("Failed to set socket permissions: %v", err)
	}

	zap.L().Sugar().Infof("CLI server listening on unix socket: %s", cliSocketPath)

	go func() {
		if err := cs.server.Serve(listener); err != nil && err != http.ErrServerClosed {
			zap.L().Sugar().Errorf("CLI server error: %v", err)
		}
	}()

	return nil
}

// Stop 停止 CLI 服务器
func (cs *CLIServer) Stop() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := cs.server.Shutdown(ctx); err != nil {
		zap.L().Sugar().Errorf("CLI server shutdown error: %v", err)
	}

	// 删除 socket 文件
	os.Remove(cliSocketPath)
}

// handleStatus 处理状态查询
func (cs *CLIServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	pid, err := ReadPID()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "daemon is not running"})
		return
	}

	status := map[string]interface{}{
		"running": true,
		"pid":     pid,
		"message": "Running",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(status)
}

// handleStop 处理停止请求
func (cs *CLIServer) handleStop(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	// 异步停止
	go func() {
		time.Sleep(100 * time.Millisecond)
		os.Exit(0)
	}()
}

// handleRestart 处理重启请求
func (cs *CLIServer) handleRestart(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	// 异步重启
	go func() {
		time.Sleep(100 * time.Millisecond)
		if err := Restart(); err != nil {
			zap.L().Sugar().Errorf("Failed to restart: %v", err)
		}
	}()
}

// handleNodes 处理节点列表请求
func (cs *CLIServer) handleNodes(w http.ResponseWriter, r *http.Request) {
	if cs.svc == nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "service not available"})
		return
	}

	nodeManager := cs.svc.GetNodeManager()
	nodes := nodeManager.ListNodes()

	result := make(map[string]interface{})
	for id, info := range nodes {
		result[id] = map[string]interface{}{
			"name":    info.Name,
			"role":    info.Role,
			"version": info.Version,
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// GetCLISocketPath 获取 CLI socket 路径
func GetCLISocketPath() string {
	return cliSocketPath
}

// handleAddHost 处理添加主机请求
func (cs *CLIServer) handleAddHost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "method not allowed"})
		return
	}

	var req services.HostProvisionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": "invalid request"})
		return
	}

	// 创建主机供应器
	provisioner, err := services.NewHostProvisioner()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": fmt.Sprintf("failed to create provisioner: %v", err)})
		return
	}

	// 执行供应
	result, err := provisioner.ProvisionHost(&req)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":  err.Error(),
			"result": result,
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}
