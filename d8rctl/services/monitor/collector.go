package monitor

import (
	"encoding/json"
	"sync"
	"time"

	"go.uber.org/zap"
)

// StatusCollector 状态收集器
type StatusCollector struct {
	mu           sync.RWMutex
	statusMap    map[string]*NodeStatus
	nodeTimeout  time.Duration
	cleanupTick  time.Duration
	stopChan     chan struct{}
}

// NewStatusCollector 创建状态收集器
func NewStatusCollector(nodeTimeout, cleanupTick time.Duration) *StatusCollector {
	c := &StatusCollector{
		statusMap:   make(map[string]*NodeStatus),
		nodeTimeout: nodeTimeout,
		cleanupTick: cleanupTick,
		stopChan:    make(chan struct{}),
	}
	go c.cleanupLoop()
	return c
}

// UpdateStatus 更新节点状态（被动接收）
func (c *StatusCollector) UpdateStatus(nodeID string, data []byte) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	var status NodeStatus
	if err := json.Unmarshal(data, &status); err != nil {
		return err
	}

	status.NodeID = nodeID
	status.LastUpdate = time.Now()
	status.Online = true

	c.statusMap[nodeID] = &status

	zap.L().Sugar().Infof("Status updated for node %s", nodeID)
	return nil
}

// GetStatus 获取节点状态
func (c *StatusCollector) GetStatus(nodeID string) (*NodeStatus, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	status, ok := c.statusMap[nodeID]
	if !ok {
		return nil, false
	}

	// 检查是否超时
	if time.Since(status.LastUpdate) > c.nodeTimeout {
		return nil, false
	}

	return status, true
}

// GetAllStatus 获取所有节点状态
func (c *StatusCollector) GetAllStatus() map[string]*NodeStatus {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]*NodeStatus)
	for nodeID, status := range c.statusMap {
		if time.Since(status.LastUpdate) <= c.nodeTimeout {
			result[nodeID] = status
		}
	}
	return result
}

// GetOnlineNodes 获取在线节点列表
func (c *StatusCollector) GetOnlineNodes() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var nodes []string
	now := time.Now()
	for nodeID, status := range c.statusMap {
		if now.Sub(status.LastUpdate) <= c.nodeTimeout {
			nodes = append(nodes, nodeID)
		}
	}
	return nodes
}

// cleanupLoop 清理循环
func (c *StatusCollector) cleanupLoop() {
	ticker := time.NewTicker(c.cleanupTick)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.stopChan:
			return
		}
	}
}

// cleanup 清理过期节点
func (c *StatusCollector) cleanup() {
	c.mu.Lock()
	defer c.mu.Unlock()

	now := time.Now()
	for nodeID, status := range c.statusMap {
		if now.Sub(status.LastUpdate) > c.nodeTimeout {
			status.Online = false
			zap.L().Sugar().Warnf("Node %s marked as offline (timeout)", nodeID)
		}
	}
}

// RemoveNode 移除节点
func (c *StatusCollector) RemoveNode(nodeID string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.statusMap, nodeID)
	zap.L().Sugar().Infof("Node %s removed from collector", nodeID)
}

// Stop 停止收集器
func (c *StatusCollector) Stop() {
	close(c.stopChan)
}