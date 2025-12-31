package monitor

import (
	"fmt"
	"time"

	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// Monitor 监控服务
type Monitor struct {
	collector *StatusCollector
}

// NewMonitor 创建监控服务
func NewMonitor() *Monitor {
	// 节点超时时间 30 秒，清理间隔 10 秒
	collector := NewStatusCollector(30*time.Second, 10*time.Second)

	return &Monitor{
		collector: collector,
	}
}

// GetCollector 获取状态收集器
func (m *Monitor) GetCollector() *StatusCollector {
	return m.collector
}

// HandleCommand 处理监控相关命令
func (m *Monitor) HandleCommand(req *pb.PublishRequest, stream pb.DomclusterService_PublishServer) *pb.PublishResponse {
	switch req.Cmd {
	case "status_update":
		// 被动接收客户端状态更新
		return HandleStatusUpdate(m.collector, req)
	case "query_response":
		// 处理客户端对查询的响应
		return HandleQueryResponse(m.collector, req)
	case "get_all_status":
		// 获取所有节点状态
		return HandleGetAllStatus(m.collector, req)
	case "get_node_status":
		// 获取单个节点状态
		return HandleGetNodeStatus(m.collector, req.Issuer, req.ReqId)
	default:
		zap.L().Sugar().Warnf("Unknown monitor command: %s", req.Cmd)
		return errorResponse(req.ReqId, "unknown command")
	}
}

// QueryNodeStatus 主动查询节点状态
func (m *Monitor) QueryNodeStatus(stream pb.DomclusterService_PublishServer) error {
	return SendStatusQuery(stream, fmt.Sprintf("query-%d", time.Now().UnixNano()))
}

// QueryAllNodeStatus 主动查询所有节点状态
func (m *Monitor) QueryAllNodeStatus(stream pb.DomclusterService_PublishServer, nodeIDs []string) error {
	for range nodeIDs {
		if err := m.QueryNodeStatus(stream); err != nil {
			zap.L().Sugar().Errorf("Failed to query node: %v", err)
		}
	}
	return nil
}

// QueryNodeResource 主动查询节点特定资源
func (m *Monitor) QueryNodeResource(stream pb.DomclusterService_PublishServer, resourceType string) error {
	return SendResourceQuery(stream, fmt.Sprintf("query-%s-%d", resourceType, time.Now().UnixNano()), resourceType)
}

// Stop 停止监控服务
func (m *Monitor) Stop() {
	m.collector.Stop()
}