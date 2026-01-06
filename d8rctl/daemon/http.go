package daemon

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"d8rctl/auth"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const httpAddr = "127.0.0.1:18080"

// ServerStatus 服务器状态
type ServerStatus struct {
	Running  bool   `json:"running"`
	PID      int    `json:"pid"`
	Uptime   string `json:"uptime"`
	Nodes    int    `json:"nodes"`
	Message  string `json:"message"`
}

// HTTPServer HTTP 服务器
type HTTPServer struct {
	server *http.Server
	status *ServerStatus
	stop   chan struct{}
	svc    interface{}
}

// NewHTTPServer 创建 HTTP 服务器
func NewHTTPServer(status *ServerStatus, svc interface{}) *HTTPServer {
	hs := &HTTPServer{
		status: status,
		stop:   make(chan struct{}),
		svc:    svc,
	}

	router := gin.Default()

	api := router.Group("/api")
	{
		api.POST("/login", hs.handleLogin)
		api.POST("/logout", auth.GinAuthMiddleware(), hs.handleLogout)
		
		authRequired := api.Group("")
		authRequired.Use(auth.GinAuthMiddleware())
		{
			authRequired.GET("/status", hs.handleStatus)
			authRequired.POST("/stop", hs.handleStop)
			authRequired.POST("/restart", hs.handleRestart)
			authRequired.GET("/docker/containers", hs.handleDockerList)
			authRequired.POST("/docker/start", hs.handleDockerStart)
			authRequired.POST("/docker/stop", hs.handleDockerStop)
			authRequired.POST("/docker/restart", hs.handleDockerRestart)
			authRequired.GET("/docker/logs", hs.handleDockerLogs)
			authRequired.GET("/docker/stats", hs.handleDockerStats)
			authRequired.GET("/docker/inspect", hs.handleDockerInspect)
			authRequired.GET("/docker/nodes", hs.handleDockerNodes)
		}
	}

	router.NoRoute(func(c *gin.Context) {
		c.File("../web-ui/dist" + c.Request.URL.Path)
	})

	hs.server = &http.Server{
		Addr:         httpAddr,
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := hs.server.Shutdown(ctx); err != nil {
		zap.L().Sugar().Errorf("HTTP server shutdown error: %v", err)
	}
}

// handleStatus 处理状态查询
func (hs *HTTPServer) handleStatus(c *gin.Context) {
	c.JSON(http.StatusOK, hs.status)
}

// handleStop 处理停止请求
func (hs *HTTPServer) handleStop(c *gin.Context) {
	hs.status.Message = "Stopping..."
	hs.status.Running = false

	c.JSON(http.StatusOK, gin.H{"status": "ok"})

	go func() {
		hs.Stop()
	}()
}

// handleRestart 处理重启请求
func (hs *HTTPServer) handleRestart(c *gin.Context) {
	hs.status.Message = "Restarting..."

	c.JSON(http.StatusOK, gin.H{"status": "ok"})

	go func() {
		hs.Stop()
	}()
}

// GetStatus 获取状态
func GetStatus() (*ServerStatus, error) {
	pid, err := ReadPID()
	if err != nil {
		return nil, fmt.Errorf("daemon is not running")
	}

	status := &ServerStatus{
		Running: true,
		PID:     pid,
		Message: "Running",
	}

	return status, nil
}

// CallStop 调用停止接口
func CallStop() error {
	resp, err := http.Post(fmt.Sprintf("http://%s/api/stop", httpAddr), "application/json", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// GetServer 获取服务实例
func (hs *HTTPServer) GetServer() interface{} {
	return hs.svc
}