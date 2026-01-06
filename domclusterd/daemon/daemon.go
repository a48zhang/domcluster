package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

	"domclusterd/config"
	"domclusterd/connections"
	"domclusterd/dockerctl"
	"domclusterd/monitor"

	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

// Daemon 守护进程
type Daemon struct {
	manager    *connections.Manager
	startTime  time.Time
	docker     *dockerctl.DockerClient
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

	// 初始化 Docker 客户端
	dockerClient, err := dockerctl.NewDockerClient()
	if err != nil {
		zap.L().Sugar().Warnf("Failed to initialize Docker client: %v (Docker features will be unavailable)", err)
		dockerClient = nil
	}

	return &Daemon{
		manager:    manager,
		startTime:  time.Now(),
		docker:     dockerClient,
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

	// 注册 Docker 处理器
	if d.docker != nil {
		dockerHandler := dockerctl.NewHandler(d.docker)
		dockerCommands := []string{"docker_list", "docker_start", "docker_stop", "docker_restart", "docker_logs", "docker_stats", "docker_inspect"}
		
		for _, cmd := range dockerCommands {
			command := cmd
			d.manager.RegisterHandler(command, func(resp *pb.PublishResponse) error {
				var data map[string]interface{}
				if err := json.Unmarshal(resp.Data, &data); err != nil {
					return err
				}
				result, err := dockerHandler.HandleCommand(command, data)
				if err != nil {
					return err
				}
				return d.manager.Send("docker_response", resp.ReqId, result)
			})
		}
		zap.L().Sugar().Info("Docker handlers registered")
	} else {
		// Docker 客户端不可用时，注册统一的错误 handler
		dockerCommands := []string{"docker_list", "docker_start", "docker_stop", "docker_restart", "docker_logs", "docker_stats", "docker_inspect"}
		for _, cmd := range dockerCommands {
			d.manager.RegisterHandler(cmd, func(resp *pb.PublishResponse) error {
				errorData := map[string]interface{}{
					"error": "Docker client not available on this node",
					"cmd":   resp.Reporter,
				}
				dataBytes, _ := json.Marshal(errorData)
				return d.manager.Send("docker_response", resp.ReqId, dataBytes)
			})
		}
		zap.L().Sugar().Warn("Docker client not available, Docker handlers registered with error responses")
	}

	zap.L().Sugar().Info("Daemon running...")
	return nil
}

// Stop 停止守护进程
func (d *Daemon) Stop() {
	zap.L().Sugar().Info("Stopping daemon...")

	// 通知控制端节点正在停止
	if d.manager != nil {
		stopData := map[string]interface{}{
			"status":  "stopping",
			"message": "Node is shutting down",
		}
		dataBytes, _ := json.Marshal(stopData)
		// 使用 PID 作为 reqID
		d.manager.Send("node_stopping", fmt.Sprintf("%d", os.Getpid()), dataBytes)
	}

	// 添加超时机制，确保优雅停止
	stopTimeout := 30 * time.Second
	stopDone := make(chan struct{})

	go func() {
		// 关闭连接
		d.manager.Close()

		// 关闭 Docker 客户端
		if d.docker != nil {
			d.docker.Close()
		}

		// 删除 PID 文件
		RemovePID()

		close(stopDone)
	}()

	// 等待停止完成或超时
	select {
	case <-stopDone:
		zap.L().Sugar().Info("Daemon stopped gracefully")
	case <-time.After(stopTimeout):
		zap.L().Sugar().Warn("Daemon stop timeout, forcing exit")
	}
}

// Restart 重启守护进程
func Restart() error {
	// 停止当前进程
	if err := Stop(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	// 轮询等待进程停止
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