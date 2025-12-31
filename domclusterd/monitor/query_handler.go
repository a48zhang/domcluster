package monitor

import (
	"context"
	"domclusterd/connections"
	"encoding/json"

	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// QueryHandler 查询处理器
type QueryHandler struct {
	monitor *Monitor
	manager *connections.Manager
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewQueryHandler 创建查询处理器
func NewQueryHandler(monitor *Monitor, manager *connections.Manager) *QueryHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &QueryHandler{
		monitor: monitor,
		manager: manager,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Register 注册查询处理器到 manager
func (qh *QueryHandler) Register() {
	qh.manager.RegisterHandler("status_query", qh.handleStatusQuery)
	qh.manager.RegisterHandler("resource_query", qh.handleResourceQuery)
	zap.L().Sugar().Info("Query handlers registered")
}

// handleStatusQuery 处理状态查询
func (qh *QueryHandler) handleStatusQuery(resp *pb.PublishResponse) error {
	report, err := qh.monitor.GetMonitorReport()
	if err != nil {
		zap.L().Sugar().Errorf("Failed to get monitor report: %v", err)
		return err
	}

	dataBytes, err := json.Marshal(report)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to marshal report: %v", err)
		return err
	}

	return qh.manager.Send("query_response", resp.ReqId, dataBytes)
}

// handleResourceQuery 处理资源查询
func (qh *QueryHandler) handleResourceQuery(resp *pb.PublishResponse) error {
	var query map[string]interface{}
	if err := json.Unmarshal(resp.Data, &query); err != nil {
		zap.L().Sugar().Errorf("Failed to unmarshal query: %v", err)
		return err
	}

	resourceType, ok := query["resource"].(string)
	if !ok {
		zap.L().Sugar().Warn("Invalid resource query: missing resource type")
		return nil
	}

	var data interface{}
	var err error

	switch resourceType {
	case "cpu":
		data, err = qh.monitor.GetSystemResources()
	case "memory":
		data, err = qh.monitor.GetSystemResources()
	case "disk":
		data, err = qh.monitor.GetSystemResources()
	case "docker":
		data, err = qh.monitor.GetDockerInfo()
	default:
		zap.L().Sugar().Warnf("Unknown resource type: %s", resourceType)
		return nil
	}

	if err != nil {
		zap.L().Sugar().Errorf("Failed to get %s info: %v", resourceType, err)
		return err
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		zap.L().Sugar().Errorf("Failed to marshal %s info: %v", resourceType, err)
		return err
	}

	return qh.manager.Send("query_response", resp.ReqId, dataBytes)
}

// Stop 停止处理器
func (qh *QueryHandler) Stop() {
	qh.cancel()
}