package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"domclusterd/config"
	"domclusterd/connections"
	"domclusterd/monitor"
	"domclusterd/tasks"
	"go.uber.org/zap"
)

// Daemon 守护进程
type Daemon struct {
	manager    *connections.Manager
	httpServer *HTTPServer
	status     *ServerStatus
	startTime  time.Time
}

// NewDaemon 创建守护进程
func NewDaemon(nodeID, nodeName string) (*Daemon, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	manager := connections.NewManager(&connections.Config{
		Address:  cfg.GetAddress(),
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		Timeout:  cfg.GetTimeout(),
	})

	// 创建 HTTP 服务器
	status := &ServerStatus{
		Running: true,
		PID:     os.Getpid(),
		Message: "Running",
		NodeID:  nodeID,
	}
	httpServer := NewHTTPServer(status)

	return &Daemon{
		manager:    manager,
		httpServer: httpServer,
		status:     status,
		startTime:  time.Now(),
	}, nil
}

// Run 运行守护进程
func (d *Daemon) Run(ctx context.Context, nodeID, nodeName string) error {
	// 写入 PID 文件
	if err := WritePID(os.Getpid()); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	defer RemovePID()

	zap.L().Sugar().Infof("Daemon started with PID: %d, NodeID: %s", os.Getpid(), nodeID)

	// 启动 HTTP 服务器
	go func() {
		if err := d.httpServer.Start(); err != nil {
			zap.L().Sugar().Error("HTTP server error", zap.Error(err))
		}
	}()

	// 更新状态
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				d.status.Uptime = time.Since(d.startTime).String()
			case <-ctx.Done():
				return
			}
		}
	}()

	// 启动连接管理器
	if err := d.manager.Start(ctx, nodeID, nodeName); err != nil {
		return fmt.Errorf("failed to start connection manager: %w", err)
	}

	// 创建监控器
	m := monitor.NewMonitor(ctx)

	// 创建并启动状态报告器（定时上报）
	reporter := monitor.NewStatusReporter(m, d.manager)
	go reporter.Start(30 * time.Second)
	defer reporter.Stop()

	// 创建并注册查询处理器
	queryHandler := monitor.NewQueryHandler(m, d.manager)
	queryHandler.Register()
	defer queryHandler.Stop()

	// 创建任务管理器
	tm := tasks.NewTaskManager(ctx)
	tm.Run()
	defer tm.Stop()

	zap.L().Sugar().Info("Daemon running...")
	return nil
}

// Stop 停止守护进程
func (d *Daemon) Stop() {
	zap.L().Sugar().Info("Stopping daemon...")
	d.status.Running = false
	d.status.Message = "Stopping"
	d.httpServer.Stop()
	d.manager.Close()
	RemovePID()
	zap.L().Sugar().Info("Daemon stopped")
}

// Restart 重启守护进程
func Restart() error {
	// 停止当前进程
	if err := CallStop(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	// 等待进程停止
	time.Sleep(2 * time.Second)

	// 重新启动
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable: %w", err)
	}

	cmd := exec.Command(executable, "daemon")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	zap.L().Sugar().Info("Daemon restarted")
	return nil
}