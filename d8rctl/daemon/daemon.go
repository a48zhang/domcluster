package daemon

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"d8rctl/auth"
	"d8rctl/connections"
	"d8rctl/services"
	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// Daemon 守护进程
type Daemon struct {
	server     *connections.Server
	httpServer *HTTPServer
	status     *ServerStatus
	startTime  time.Time
}

// NewDaemon 创建守护进程
func NewDaemon() (*Daemon, error) {
	if err := auth.GetPasswordManager().Init(); err != nil {
		return nil, fmt.Errorf("failed to initialize password manager: %w", err)
	}

	config := &connections.Config{
		Address:  ":50051",
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
	}

	server, err := connections.NewServer(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	domclusterServer := services.NewDomclusterServer()
	pb.RegisterDomclusterServiceServer(server.GetServer(), domclusterServer)

	status := &ServerStatus{
		Running: true,
		PID:     os.Getpid(),
		Message: "Running",
	}
	httpServer := NewHTTPServer(status, domclusterServer)

	return &Daemon{
		server:     server,
		httpServer: httpServer,
		status:     status,
		startTime:  time.Now(),
	}, nil
}

// Run 运行守护进程
func (d *Daemon) Run(ctx context.Context) error {
	if err := WritePID(os.Getpid()); err != nil {
		return fmt.Errorf("failed to write PID file: %w", err)
	}
	defer RemovePID()

	zap.L().Sugar().Infof("Daemon started with PID: %d", os.Getpid())

	go func() {
		if err := d.httpServer.Start(); err != nil {
			zap.L().Sugar().Error("HTTP server error", zap.Error(err))
		}
	}()

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

	zap.L().Sugar().Info("Starting gRPC server...")
	if err := d.server.Start(ctx); err != nil {
		return fmt.Errorf("gRPC server error: %w", err)
	}

	return nil
}

// Stop 停止守护进程
func (d *Daemon) Stop() {
	zap.L().Sugar().Info("Stopping daemon...")
	d.status.Running = false
	d.status.Message = "Stopping"
	d.httpServer.Stop()
	d.server.Stop()
	RemovePID()
	zap.L().Sugar().Info("Daemon stopped")
}

// Restart 重启守护进程
func Restart() error {
	if err := CallStop(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	const maxWaitTime = 30 * time.Second
	const checkInterval = 500 * time.Millisecond
	startTime := time.Now()

	for {
		if !IsRunning() {
			zap.L().Sugar().Info("Daemon stopped successfully")
			break
		}

		elapsed := time.Since(startTime)
		if elapsed >= maxWaitTime {
			zap.L().Sugar().Warnf("Daemon did not stop within %v, proceeding with restart", maxWaitTime)
			break
		}

		time.Sleep(checkInterval)
	}

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