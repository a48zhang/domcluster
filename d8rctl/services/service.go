package services

import (
	"d8rctl/services/monitor"
	"fmt"
	pb "domcluster/api/proto"
	"go.uber.org/zap"
	"sync"
)

// DomclusterServer Domcluster 服务端
type DomclusterServer struct {
	pb.UnimplementedDomclusterServiceServer
	nodeManager      *NodeManager
	monitor          *monitor.Monitor
	dockerResponses  map[string]chan *DockerResult
	dockerResponsesMu sync.RWMutex
	streams          map[string]pb.DomclusterService_PublishServer
	streamsMu        sync.RWMutex
}

// NewDomclusterServer 创建 Domcluster 服务端
func NewDomclusterServer() *DomclusterServer {
	return &DomclusterServer{
		nodeManager:     NewNodeManager(),
		monitor:         monitor.NewMonitor(),
		dockerResponses: make(map[string]chan *DockerResult),
		streams:         make(map[string]pb.DomclusterService_PublishServer),
	}
}

// Publish 处理发布流
func (s *DomclusterServer) Publish(stream pb.DomclusterService_PublishServer) error {
	var currentIssuer string

	for {
		req, err := stream.Recv()
		if err != nil {
			zap.L().Sugar().Errorf("Publish recv error: %v", err)
			// 清理 streams map 中的条目
			if currentIssuer != "" {
				s.streamsMu.Lock()
				delete(s.streams, currentIssuer)
				s.streamsMu.Unlock()
				zap.L().Sugar().Infof("Removed stream for issuer: %s", currentIssuer)
			}
			return err
		}

		zap.L().Sugar().Debugf("Received: issuer=%s, req_id=%s, cmd=%s", req.Issuer, req.ReqId, req.Cmd)

		currentIssuer = req.Issuer

		// 保存 stream 引用
		s.streamsMu.Lock()
		s.streams[req.Issuer] = stream
		s.streamsMu.Unlock()

		// 检查是否是 Docker 响应
		if req.Cmd == "docker_response" {
			s.handleDockerResponse(req)
			continue
		}

		// 根据命令类型处理
		resp := s.handleRequest(req)

		// 发送回复
		if err := stream.Send(resp); err != nil {
			zap.L().Sugar().Errorf("Publish send error: %v", err)
			// 清理 streams map 中的条目
			s.streamsMu.Lock()
			delete(s.streams, req.Issuer)
			s.streamsMu.Unlock()
			zap.L().Sugar().Infof("Removed stream for issuer: %s due to send error", req.Issuer)
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

// RegisterDockerResponse 注册 Docker 响应处理器
func (s *DomclusterServer) RegisterDockerResponse(reqID string, resultChan chan *DockerResult) {
	s.dockerResponsesMu.Lock()
	defer s.dockerResponsesMu.Unlock()
	s.dockerResponses[reqID] = resultChan
}

// handleDockerResponse 处理 Docker 响应
func (s *DomclusterServer) handleDockerResponse(req *pb.PublishRequest) {
	s.dockerResponsesMu.RLock()
	resultChan, ok := s.dockerResponses[req.ReqId]
	s.dockerResponsesMu.RUnlock()

	if ok {
		resultChan <- &DockerResult{
			Status: 0,
			Data:   req.Data,
		}
		s.dockerResponsesMu.Lock()
		delete(s.dockerResponses, req.ReqId)
		s.dockerResponsesMu.Unlock()
	}
}

// SendToNode 发送命令到指定节点
func (s *DomclusterServer) SendToNode(nodeID, cmd, reqID string, data []byte) error {
	s.streamsMu.RLock()
	stream, ok := s.streams[nodeID]
	s.streamsMu.RUnlock()

	if !ok {
		return fmt.Errorf("node %s not connected", nodeID)
	}

	resp := &pb.PublishResponse{
		Reporter: "server",
		ReqId:    reqID,
		Status:   0,
		Data:     data,
	}

	return stream.Send(resp)
}