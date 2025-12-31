package daemon

import (
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

const httpAddr = "127.0.0.1:18081"

// ServerStatus 服务器状态
type ServerStatus struct {
	Running  bool   `json:"running"`
	PID      int    `json:"pid"`
	Uptime   string `json:"uptime"`
	Message  string `json:"message"`
	NodeID   string `json:"node_id"`
}

// HTTPServer HTTP 服务器
type HTTPServer struct {
	server *http.Server
	status *ServerStatus
	stop   chan struct{}
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(status *ServerStatus) *HTTPServer {
	mux := http.NewServeMux()
	hs := &HTTPServer{
		status: status,
		stop:   make(chan struct{}),
	}

	mux.HandleFunc("/status", hs.handleStatus)
	mux.HandleFunc("/stop", hs.handleStop)
	mux.HandleFunc("/restart", hs.handleRestart)

	hs.server = &http.Server{
		Addr:    httpAddr,
		Handler: mux,
	}

	return hs
}

// Start 启动 HTTP 服务器
func (hs *HTTPServer) Start() error {
	zap.L().Sugar().Infof("HTTP server listening on %s", httpAddr)
	return hs.server.ListenAndServe()
}

// Stop 停止 HTTP 服务器
func (hs *HTTPServer) Stop() {
	close(hs.stop)
	hs.server.Close()
}

// handleStatus 处理状态查询
func (hs *HTTPServer) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(hs.status)
}

// handleStop 处理停止请求
func (hs *HTTPServer) handleStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hs.status.Message = "Stopping..."
	hs.status.Running = false

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	// 触发停止
	go func() {
		hs.Stop()
	}()
}

// handleRestart 处理重启请求
func (hs *HTTPServer) handleRestart(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	hs.status.Message = "Restarting..."

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})

	// 触发重启
	go func() {
		hs.Stop()
		// 重启由外部进程处理
	}()
}

// GetStatus 获取状态
func GetStatus() (*ServerStatus, error) {
	resp, err := http.Get(fmt.Sprintf("http://%s/status", httpAddr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var status ServerStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, err
	}

	return &status, nil
}

// CallStop 调用停止接口
func CallStop() error {
	resp, err := http.Post(fmt.Sprintf("http://%s/stop", httpAddr), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return err
	}

	if result["status"] != "ok" {
		return fmt.Errorf("failed to stop")
	}

	return nil
}