package services

import (
	"d8rctl/services/monitor"
	"encoding/json"

	pb "domcluster/api/proto"

	"go.uber.org/zap"
)

// handleRequest 处理请求
func (s *DomclusterServer) handleRequest(req *pb.PublishRequest) *pb.PublishResponse {
	switch req.Cmd {
	case "register":
		return s.handleRegister(req)
	case "heartbeat":
		return s.handleHeartbeat(req)
	case "command_result":
		return s.handleCommandResult(req)
	case "command_output":
		return s.handleCommandOutput(req)
	case "status_update":
		return s.handleStatusUpdate(req)
	default:
		return &pb.PublishResponse{
			Reporter: "server",
			ReqId:    req.ReqId,
			Status:   -1,
			Data:     []byte(`{"error":"unknown command"}`),
		}
	}
}

// handleRegister 处理节点注册请求
func (s *DomclusterServer) handleRegister(req *pb.PublishRequest) *pb.PublishResponse {
	var data map[string]interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return errorResponse(req.ReqId, "invalid data")
	}

	name, ok := data["name"].(string)
	if !ok {
		zap.L().Sugar().Errorf("Type assertion failed for 'name' field in node registration request from %s", req.Issuer)
		return errorResponse(req.ReqId, "invalid name field")
	}

	version, ok := data["version"].(string)
	if !ok {
		zap.L().Sugar().Errorf("Type assertion failed for 'version' field in node registration request from %s", req.Issuer)
		return errorResponse(req.ReqId, "invalid version field")
	}

	s.nodeManager.AddNode(req.Issuer, &NodeInfo{
		Name:    name,
		Role:    "worker", // 默认角色，后续可由 ctl 分配
		Version: version,
	})

	zap.L().Sugar().Infof("Node registered: %s (%s)", name, req.Issuer)

	return successResponse(req.ReqId, map[string]interface{}{
		"message": "registered",
	})
}

// handleHeartbeat 处理心跳请求
func (s *DomclusterServer) handleHeartbeat(req *pb.PublishRequest) *pb.PublishResponse {
	if _, ok := s.nodeManager.GetNode(req.Issuer); !ok {
		return errorResponse(req.ReqId, "node not registered")
	}

	return successResponse(req.ReqId, map[string]interface{}{
		"timestamp": req.Data,
	})
}

// handleCommandResult 处理命令结果
func (s *DomclusterServer) handleCommandResult(req *pb.PublishRequest) *pb.PublishResponse {
	zap.L().Sugar().Infof("Command result from %s", req.Issuer)
	return successResponse(req.ReqId, nil)
}

// handleCommandOutput 处理命令输出
func (s *DomclusterServer) handleCommandOutput(req *pb.PublishRequest) *pb.PublishResponse {
	var data map[string]interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		zap.L().Sugar().Errorf("Failed to unmarshal command output: %v", err)
		return errorResponse(req.ReqId, "invalid data")
	}

	outputType, _ := data["type"].(string)
	output, _ := data["output"].(string)

	zap.L().Sugar().Infof("[%s] %s", outputType, output)
	return successResponse(req.ReqId, nil)
}

// handleStatusUpdate 处理状态更新请求
func (s *DomclusterServer) handleStatusUpdate(req *pb.PublishRequest) *pb.PublishResponse {
	if _, ok := s.nodeManager.GetNode(req.Issuer); !ok {
		return errorResponse(req.ReqId, "node not registered")
	}

	// 使用监控服务处理状态更新
	return monitor.HandleStatusUpdate(s.monitor.GetCollector(), req)
}
