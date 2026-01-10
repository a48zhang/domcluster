package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"time"

	"d8rctl/auth"
	"d8rctl/services"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

const httpAddr = ":18080"

// ServerStatus 服务器状态
type ServerStatus struct {
	Running bool   `json:"running"`
	PID     int    `json:"pid"`
	Uptime  string `json:"uptime"`
	Nodes   int    `json:"nodes"`
	Message string `json:"message"`
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

	// 添加 CORS 中间件
	router.Use(func(c *gin.Context) {
		// 获取请求的origin
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		}
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	api := router.Group("/api")
	{
		// Web UI 端点 - 需要认证
		api.POST("/login", hs.handleLogin)
		api.POST("/logout", auth.GinAuthMiddleware(), hs.handleLogout)

		authRequired := api.Group("")
		authRequired.Use(auth.GinAuthMiddleware())
		{
			authRequired.GET("/status", hs.handleStatus)
			authRequired.POST("/stop", hs.handleStop)
			authRequired.POST("/restart", hs.handleRestart)
			authRequired.GET("/nodes", hs.handleNodes)
			authRequired.GET("/nodes/:nodeId/status", hs.handleNodeStatus)
			authRequired.POST("/hosts/add", hs.handleAddHost)
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
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", cliSocketPath)
			},
		},
	}

	resp, err := client.Post("http://unix/stop", "application/json", nil)
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

// handleNodes 处理节点列表请求
func (hs *HTTPServer) handleNodes(c *gin.Context) {
	if hs.svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}

	domclusterServer, ok := hs.svc.(*services.DomclusterServer)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid service type"})
		return
	}

	nodeManager := domclusterServer.GetNodeManager()
	nodes := nodeManager.ListNodes()

	result := make(map[string]interface{})
	for id, info := range nodes {
		result[id] = map[string]interface{}{
			"name":    info.Name,
			"role":    info.Role,
			"version": info.Version,
		}
	}

	c.JSON(http.StatusOK, result)
}

// handleNodeStatus 处理获取单个节点状态请求
func (hs *HTTPServer) handleNodeStatus(c *gin.Context) {
	if hs.svc == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "service not available"})
		return
	}

	domclusterServer, ok := hs.svc.(*services.DomclusterServer)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "invalid service type"})
		return
	}

	nodeID := c.Param("nodeId")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node ID is required"})
		return
	}

	monitor := domclusterServer.GetMonitor()
	if monitor == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "monitor not available"})
		return
	}

	collector := monitor.GetCollector()
	status, exists := collector.GetStatus(nodeID)
	if !exists {
		c.JSON(http.StatusNotFound, gin.H{"error": "node not found"})
		return
	}

	c.JSON(http.StatusOK, status)
}

// GetNodeList 获取节点列表
func GetNodeList() (map[string]interface{}, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", cliSocketPath)
			},
		},
	}

	resp, err := client.Get("http://unix/nodes")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get node list, status: %d", resp.StatusCode)
	}

	var nodes map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&nodes); err != nil {
		return nil, err
	}

	return nodes, nil
}

// handleAddHost 处理添加主机请求
func (hs *HTTPServer) handleAddHost(c *gin.Context) {
	var req services.HostProvisionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// 创建主机供应器
	provisioner, err := services.NewHostProvisioner()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("failed to create provisioner: %v", err)})
		return
	}

	// 执行供应
	result, err := provisioner.ProvisionHost(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":  err.Error(),
			"result": result,
		})
		return
	}

	c.JSON(http.StatusOK, result)
}
