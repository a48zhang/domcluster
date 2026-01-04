package services

import (
	"encoding/json"

	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// successResponse 创建成功响应
func successResponse(reqID string, data interface{}) *pb.PublishResponse {
	var dataBytes []byte
	if data != nil {
		var err error
		dataBytes, err = json.Marshal(data)
		if err != nil {
			zap.L().Sugar().Errorf("Failed to marshal success response data (reqID: %s): %v", reqID, err)
			// 使用默认值，避免返回空数据
			dataBytes = []byte(`{"message":"internal error: failed to marshal response"}`)
		}
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
	dataBytes, err := json.Marshal(map[string]string{"error": errMsg})
	if err != nil {
		zap.L().Sugar().Errorf("Failed to marshal error response (reqID: %s, errMsg: %s): %v", reqID, errMsg, err)
		// 使用默认错误消息，确保客户端能收到错误信息
		dataBytes = []byte(`{"error":"internal error"}`)
	}
	return &pb.PublishResponse{
		Reporter: "server",
		ReqId:    reqID,
		Status:   -1,
		Data:     dataBytes,
	}
}