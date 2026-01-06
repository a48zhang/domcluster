package cli

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"go.uber.org/zap"
)

// Start 启动守护进程
func Start() error {
	if IsRunning() {
		return fmt.Errorf("daemon is already running")
	}

	logger, err := zap.NewProduction()
	if err != nil {
		return err
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable: %w", err)
	}

	cmd := exec.Command(executable, "daemon")
	cmd.Stdout = nil
	cmd.Stderr = nil
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