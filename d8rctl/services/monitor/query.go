package monitor

import (
	"encoding/json"
	"fmt"
	"time"

	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// SendStatusQuery 发送状态查询请求
func SendStatusQuery(stream pb.DomclusterService_PublishServer, reqID string) error {
	data := map[string]interface{}{
		"cmd":       "status_query",
		"timestamp": time.Now().Unix(),
	}
	dataBytes, _ := json.Marshal(data)

	resp := &pb.PublishResponse{
		Reporter: "server",
		ReqId:    reqID,
		Status:   0,
		Data:     dataBytes,
	}

	if err := stream.Send(resp); err != nil {
		return fmt.Errorf("failed to send status query: %w", err)
	}
	zap.L().Sugar().Info("Status query request sent")
	return nil
}

// SendResourceQuery 发送资源查询请求
func SendResourceQuery(stream pb.DomclusterService_PublishServer, reqID string, resourceType string) error {
	data := map[string]interface{}{
		"cmd":         "resource_query",
		"resource":    resourceType,
		"timestamp":   time.Now().Unix(),
	}
	dataBytes, _ := json.Marshal(data)

	resp := &pb.PublishResponse{
		Reporter: "server",
		ReqId:    reqID,
		Status:   0,
		Data:     dataBytes,
	}

	if err := stream.Send(resp); err != nil {
		return fmt.Errorf("failed to send resource query: %w", err)
	}
	zap.L().Sugar().Infof("Resource query %s request sent", resourceType)
	return nil
}