package services

import (
	"encoding/json"
	"log"

	pb "domcluster/api/proto"
)

type DomclusterServer struct {
	pb.UnimplementedDomclusterServiceServer
	nodes map[string]*NodeInfo
}

type NodeInfo struct {
	Name    string
	Role    string
	Version string
}

func NewDomclusterServer() *DomclusterServer {
	return &DomclusterServer{
		nodes: make(map[string]*NodeInfo),
	}
}

func (s *DomclusterServer) Publish(stream pb.DomclusterService_PublishServer) error {
	for {
		req, err := stream.Recv()
		if err != nil {
			log.Printf("Publish recv error: %v", err)
			return err
		}

		log.Printf("Received: issuer=%s, req_id=%s, cmd=%s", req.Issuer, req.ReqId, req.Cmd)

		// 根据命令类型处理
		resp := s.handleRequest(req)

		// 发送回复
		if err := stream.Send(resp); err != nil {
			log.Printf("Publish send error: %v", err)
			return err
		}
	}
}

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
	default:
		return &pb.PublishResponse{
			Reporter: "server",
			ReqId:    req.ReqId,
			Status:   -1,
			Data:     []byte(`{"error":"unknown command"}`),
		}
	}
}

func (s *DomclusterServer) handleRegister(req *pb.PublishRequest) *pb.PublishResponse {
	var data map[string]interface{}
	if err := json.Unmarshal(req.Data, &data); err != nil {
		return errorResponse(req.ReqId, "invalid data")
	}

	name := data["name"].(string)
	role := data["role"].(string)
	version := data["version"].(string)

	s.nodes[req.Issuer] = &NodeInfo{
		Name:    name,
		Role:    role,
		Version: version,
	}

	log.Printf("Node registered: %s (%s)", name, req.Issuer)

	return successResponse(req.ReqId, map[string]interface{}{
		"message": "registered",
	})
}

func (s *DomclusterServer) handleHeartbeat(req *pb.PublishRequest) *pb.PublishResponse {
	if _, ok := s.nodes[req.Issuer]; !ok {
		return errorResponse(req.ReqId, "node not registered")
	}

	return successResponse(req.ReqId, map[string]interface{}{
		"timestamp": req.Data,
	})
}

func (s *DomclusterServer) handleCommandResult(req *pb.PublishRequest) *pb.PublishResponse {
	log.Printf("Command result from %s", req.Issuer)
	return successResponse(req.ReqId, nil)
}

func (s *DomclusterServer) handleCommandOutput(req *pb.PublishRequest) *pb.PublishResponse {
	var data map[string]interface{}
	json.Unmarshal(req.Data, &data)
	outputType := data["type"]
	output := data["output"]

	log.Printf("[%s] %s", outputType, output)
	return successResponse(req.ReqId, nil)
}

func successResponse(reqID string, data interface{}) *pb.PublishResponse {
	var dataBytes []byte
	if data != nil {
		dataBytes, _ = json.Marshal(data)
	}
	return &pb.PublishResponse{
		Reporter: "server",
		ReqId:    reqID,
		Status:   0,
		Data:     dataBytes,
	}
}

func errorResponse(reqID string, errMsg string) *pb.PublishResponse {
	dataBytes, _ := json.Marshal(map[string]string{"error": errMsg})
	return &pb.PublishResponse{
		Reporter: "server",
		ReqId:    reqID,
		Status:   -1,
		Data:     dataBytes,
	}
}