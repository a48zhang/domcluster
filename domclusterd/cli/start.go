package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"go.uber.org/zap"
)

// Start 启动守护进程
func Start(nodeID, nodeName string) error {
	// 检查是否已经在运行
	if IsRunning() {
		return fmt.Errorf("daemon is already running")
	}

	// 初始化日志
	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// 获取可执行文件路径
	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable: %w", err)
	}

	// 启动守护进程
	cmd := exec.Command(executable, "daemon", nodeID, nodeName)
	cmd.Stdout = nil   // 不输出到终端
	cmd.Stderr = nil   // 不输出到终端
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true, // Windows: 隐藏窗口
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("Daemon started with PID: %d\n", cmd.Process.Pid)
	return nil
}