package services

import (
	"d8rctl/services/monitor"
	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// DomclusterServer Domcluster 服务端
type DomclusterServer struct {
	pb.UnimplementedDomclusterServiceServer
	nodeManager *NodeManager
	monitor     *monitor.Monitor
}

// NewDomclusterServer 创建 Domcluster 服务端
func NewDomclusterServer() *DomclusterServer {
	return &DomclusterServer{
		nodeManager: NewNodeManager(),
		monitor:     monitor.NewMonitor(),
	}
}

// Publish 处理发布流
func (s *DomclusterServer) Publish(stream pb.DomclusterService_PublishServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			zap.L().Sugar().Errorf("Publish recv error: %v", err)
			return err
		}

		zap.L().Sugar().Debugf("Received: issuer=%s, req_id=%s, cmd=%s", req.Issuer, req.ReqId, req.Cmd)

		// 根据命令类型处理
		resp := s.handleRequest(req)

		// 发送回复
		if err := stream.Send(resp); err != nil {
			zap.L().Sugar().Errorf("Publish send error: %v", err)
			return err
		}
	}
}

// GetNodeManager 获取节点管理器
func (s *DomclusterServer) GetNodeManager() *NodeManager {
	return s.nodeManager
}

// GetMonitor 获取监控服务
func (s *DomclusterServer) GetMonitor() *monitor.Monitor {
	return s.monitor
}