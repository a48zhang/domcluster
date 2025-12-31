package services

// NodeInfo 节点信息
type NodeInfo struct {
	Name    string
	Role    string
	Version string
}

// NodeManager 节点管理器
type NodeManager struct {
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
	nm.nodes[nodeID] = info
}

// GetNode 获取节点信息
func (nm *NodeManager) GetNode(nodeID string) (*NodeInfo, bool) {
	info, ok := nm.nodes[nodeID]
	return info, ok
}

// RemoveNode 移除节点
func (nm *NodeManager) RemoveNode(nodeID string) {
	delete(nm.nodes, nodeID)
}

// ListNodes 列出所有节点
func (nm *NodeManager) ListNodes() map[string]*NodeInfo {
	return nm.nodes
}