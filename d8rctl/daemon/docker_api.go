package daemon

import (
	"context"
	"net/http"
	"time"

	"d8rctl/services"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultRequestTimeout 默认请求超时时间
	DefaultRequestTimeout = 60 * time.Second
)

// handleDockerList 处理列出容器请求
func (hs *HTTPServer) handleDockerList(c *gin.Context) {
	nodeID := c.Query("node_id")
	if nodeID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id is required"})
		return
	}

	allStr := c.Query("all")
	all := allStr == "true" || allStr == "1"

	ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.ListContainers(ctx, nodeID, all)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDockerStart 处理启动容器请求
func (hs *HTTPServer) handleDockerStart(c *gin.Context) {
	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id and container_id are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.StartContainer(ctx, req.NodeID, req.ContainerID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDockerStop 处理停止容器请求
func (hs *HTTPServer) handleDockerStop(c *gin.Context) {
	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
		Timeout     int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id and container_id are required"})
		return
	}

	if req.Timeout == 0 {
		req.Timeout = 10
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.StopContainer(ctx, req.NodeID, req.ContainerID, req.Timeout)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDockerRestart 处理重启容器请求
func (hs *HTTPServer) handleDockerRestart(c *gin.Context) {
	var req struct {
		NodeID      string `json:"node_id"`
		ContainerID string `json:"container_id"`
		Timeout     int    `json:"timeout"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	if req.NodeID == "" || req.ContainerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id and container_id are required"})
		return
	}

	if req.Timeout == 0 {
		req.Timeout = 10
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.RestartContainer(ctx, req.NodeID, req.ContainerID, req.Timeout)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDockerLogs 处理获取容器日志请求
func (hs *HTTPServer) handleDockerLogs(c *gin.Context) {
	nodeID := c.Query("node_id")
	containerID := c.Query("container_id")
	tail := c.Query("tail")

	if nodeID == "" || containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id and container_id are required"})
		return
	}

	if tail == "" {
		tail = "100"
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.GetContainerLogs(ctx, nodeID, containerID, tail)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDockerStats 处理获取容器统计信息请求
func (hs *HTTPServer) handleDockerStats(c *gin.Context) {
	nodeID := c.Query("node_id")
	containerID := c.Query("container_id")

	if nodeID == "" || containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id and container_id are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.GetContainerStats(ctx, nodeID, containerID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDockerInspect 处理查看容器详情请求
func (hs *HTTPServer) handleDockerInspect(c *gin.Context) {
	nodeID := c.Query("node_id")
	containerID := c.Query("container_id")

	if nodeID == "" || containerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node_id and container_id are required"})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), DefaultRequestTimeout)
	defer cancel()

	dockerHandler := services.NewDockerHandler(hs.svc.(*services.DomclusterServer))
	result, err := dockerHandler.InspectContainer(ctx, nodeID, containerID)
	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			c.JSON(http.StatusGatewayTimeout, gin.H{"error": "request timeout"})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, result)
}

// handleDockerNodes 处理获取所有节点列表
func (hs *HTTPServer) handleDockerNodes(c *gin.Context) {
	nodes := hs.svc.(*services.DomclusterServer).GetNodeManager().ListNodes()
	c.JSON(http.StatusOK, gin.H{"nodes": nodes})
}