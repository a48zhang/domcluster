package monitor

import (
	"context"
	"domclusterd/connections"
	"encoding/json"
	"fmt"
	"time"

	"go.uber.org/zap"
)

// StatusReporter 状态报告器
type StatusReporter struct {
	monitor *Monitor
	manager *connections.Manager
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewStatusReporter 创建状态报告器
func NewStatusReporter(monitor *Monitor, manager *connections.Manager) *StatusReporter {
	ctx, cancel := context.WithCancel(context.Background())
	return &StatusReporter{
		monitor: monitor,
		manager: manager,
		ctx:     ctx,
		cancel:  cancel,
	}
}

// Start 启动定时上报协程
func (sr *StatusReporter) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	zap.L().Sugar().Infof("Starting status reporter with interval: %v", interval)

	for {
		select {
		case <-ticker.C:
			if err := sr.reportStatus(); err != nil {
				zap.L().Sugar().Errorf("Failed to report status: %v", err)
			}
		case <-sr.ctx.Done():
			zap.L().Sugar().Info("Status reporter stopped")
			return
		}
	}
}

// reportStatus 上报状态
func (sr *StatusReporter) reportStatus() error {
	report, err := sr.monitor.GetMonitorReport()
	if err != nil {
		return fmt.Errorf("failed to get monitor report: %w", err)
	}

	dataBytes, err := json.Marshal(report)
	if err != nil {
		return fmt.Errorf("failed to marshal report: %w", err)
	}

	reqID := fmt.Sprintf("status-%d", time.Now().UnixNano())
	return sr.manager.Send("status_update", reqID, dataBytes)
}

// Stop 停止报告器
func (sr *StatusReporter) Stop() {
	sr.cancel()
}