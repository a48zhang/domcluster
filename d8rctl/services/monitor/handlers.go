package monitor

import (
	"encoding/json"

	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// HandleStatusUpdate 处理状态更新请求（被动接收）
func HandleStatusUpdate(collector *StatusCollector, req *pb.PublishRequest) *pb.PublishResponse {
	if err := collector.UpdateStatus(req.Issuer, req.Data); err != nil {
		zap.L().Sugar().Errorf("Failed to update status for node %s: %v", req.Issuer, err)
		return errorResponse(req.ReqId, "failed to update status")
	}

	zap.L().Sugar().Infof("Status updated successfully for node %s", req.Issuer)
	zap.L().Sugar().Debugf("Updated status data: %s", string(req.Data))
	return successResponse(req.ReqId, map[string]interface{}{
		"message": "status updated",
	})
}

// HandleQueryResponse 处理查询响应（客户端响应查询请求）
func HandleQueryResponse(collector *StatusCollector, req *pb.PublishRequest) *pb.PublishResponse {
	// 客户端响应查询请求，更新状态
	if err := collector.UpdateStatus(req.Issuer, req.Data); err != nil {
		zap.L().Sugar().Errorf("Failed to process query response from node %s: %v", req.Issuer, err)
		return errorResponse(req.ReqId, "failed to process response")
	}

	zap.L().Sugar().Infof("Query response received from node %s", req.Issuer)
	return successResponse(req.ReqId, map[string]interface{}{
		"message": "response received",
	})
}

// HandleGetAllStatus 处理获取所有状态请求
func HandleGetAllStatus(collector *StatusCollector, req *pb.PublishRequest) *pb.PublishResponse {
	statusMap := collector.GetAllStatus()

	return successResponse(req.ReqId, map[string]interface{}{
		"nodes":      len(statusMap),
		"node_list":  collector.GetOnlineNodes(),
		"statuses":   statusMap,
	})
}

// HandleGetNodeStatus 处理获取单个节点状态请求
func HandleGetNodeStatus(collector *StatusCollector, nodeID string, reqID string) *pb.PublishResponse {
	status, ok := collector.GetStatus(nodeID)
	if !ok {
		return errorResponse(reqID, "node not found or offline")
	}

	return successResponse(reqID, status)
}

// successResponse 创建成功响应
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

// errorResponse 创建错误响应
func errorResponse(reqID string, errMsg string) *pb.PublishResponse {
	dataBytes, _ := json.Marshal(map[string]string{"error": errMsg})
	return &pb.PublishResponse{
		Reporter: "server",
		ReqId:    reqID,
		Status:   -1,
		Data:     dataBytes,
	}
}