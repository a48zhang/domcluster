package services

import "sync"

// NodeInfo 节点信息
type NodeInfo struct {
	Name    string
	Role    string
	Version string
}

// NodeManager 节点管理器
type NodeManager struct {
	mu    sync.RWMutex
	nodes map[string]*NodeInfo
}

// NewNodeManager 创建节点管理器
func NewNodeManager() *NodeManager {
	return &NodeManager{
		nodes: make(map[string]*NodeInfo),
	}
}

// AddNode 添加节点
func (nm *NodeManager) AddNode(nodeID string, info *NodeInfo) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	nm.nodes[nodeID] = info
}

// GetNode 获取节点信息
func (nm *NodeManager) GetNode(nodeID string) (*NodeInfo, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	info, ok := nm.nodes[nodeID]
	return info, ok
}

// RemoveNode 移除节点
func (nm *NodeManager) RemoveNode(nodeID string) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	delete(nm.nodes, nodeID)
}

// ListNodes 列出所有节点
func (nm *NodeManager) ListNodes() map[string]*NodeInfo {
	nm.mu.RLock()
	defer nm.mu.RUnlock()

	// 返回副本以避免外部修改导致的并发问题
	result := make(map[string]*NodeInfo, len(nm.nodes))
	for id, info := range nm.nodes {
		result[id] = info
	}
	return result
}