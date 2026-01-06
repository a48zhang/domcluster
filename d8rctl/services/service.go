package services

import (
	"d8rctl/services/monitor"
	"fmt"
	pb "domcluster/api/proto"
	"go.uber.org/zap"
	"sync"
	"time"
)

// DomclusterServer Domcluster 服务端
type DomclusterServer struct {
	pb.UnimplementedDomclusterServiceServer
	nodeManager              *NodeManager
	monitor                  *monitor.Monitor
	dockerResponses          map[string]chan *DockerResult
	dockerResponseTimestamps map[string]time.Time
	dockerResponsesMu        sync.RWMutex
	streams                  map[string]pb.DomclusterService_PublishServer
	streamsMu                sync.RWMutex
	cleanupDone              chan struct{}
}

// NewDomclusterServer 创建 Domcluster 服务端
func NewDomclusterServer() *DomclusterServer {
	s := &DomclusterServer{
		nodeManager:              NewNodeManager(),
		monitor:                  monitor.NewMonitor(),
		dockerResponses:          make(map[string]chan *DockerResult),
		dockerResponseTimestamps: make(map[string]time.Time),
		streams:                  make(map[string]pb.DomclusterService_PublishServer),
		cleanupDone:              make(chan struct{}),
	}
	go s.cleanupExpiredResponses()
	return s
}

// Publish 处理发布流
func (s *DomclusterServer) Publish(stream pb.DomclusterService_PublishServer) error {
	var currentIssuer string

	for {
		req, err := stream.Recv()
		if err != nil {
			zap.L().Sugar().Errorf("Publish recv error: %v", err)
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

		s.streamsMu.Lock()
		s.streams[req.Issuer] = stream
		s.streamsMu.Unlock()

		if req.Cmd == "docker_response" {
			s.handleDockerResponse(req)
			continue
		}

		resp := s.handleRequest(req)

		if err := stream.Send(resp); err != nil {
			zap.L().Sugar().Errorf("Publish send error: %v", err)
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
	s.dockerResponseTimestamps[reqID] = time.Now()
}

// handleDockerResponse 处理 Docker 响应
func (s *DomclusterServer) handleDockerResponse(req *pb.PublishRequest) {
	s.dockerResponsesMu.RLock()
	resultChan, ok := s.dockerResponses[req.ReqId]
	s.dockerResponsesMu.RUnlock()

	if ok {
		select {
		case resultChan <- &DockerResult{
			Status: 0,
			Data:   req.Data,
		}:
			s.dockerResponsesMu.Lock()
			delete(s.dockerResponses, req.ReqId)
			delete(s.dockerResponseTimestamps, req.ReqId)
			s.dockerResponsesMu.Unlock()
		default:
			// channel 已关闭或缓冲区已满
			zap.L().Sugar().Warnf("Failed to send docker response for reqID: %s, channel may be closed", req.ReqId)
			s.dockerResponsesMu.Lock()
			delete(s.dockerResponses, req.ReqId)
			delete(s.dockerResponseTimestamps, req.ReqId)
			s.dockerResponsesMu.Unlock()
		}
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

// cleanupExpiredResponses 定期清理过期的 docker response channel
func (s *DomclusterServer) cleanupExpiredResponses() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			s.cleanupOldResponses()
		case <-s.cleanupDone:
			return
		}
	}
}

// cleanupOldResponses 清理超过 5 分钟的响应 channel
func (s *DomclusterServer) cleanupOldResponses() {
	s.dockerResponsesMu.Lock()
	defer s.dockerResponsesMu.Unlock()

	now := time.Now()
	expiryDuration := 5 * time.Minute

	for reqID, timestamp := range s.dockerResponseTimestamps {
		if now.Sub(timestamp) > expiryDuration {
			delete(s.dockerResponses, reqID)
			delete(s.dockerResponseTimestamps, reqID)
			zap.L().Sugar().Infof("Cleaned up expired docker response for reqID: %s", reqID)
		}
	}
}

// Shutdown 关闭服务器，停止清理 goroutine
func (s *DomclusterServer) Shutdown() {
	if s.cleanupDone != nil {
		close(s.cleanupDone)
	}
}