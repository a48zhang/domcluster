package services

import (
	"encoding/json"

	pb "domcluster/api/proto"
)

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