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

	logFile, err := os.OpenFile("d8rctl.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	cmd := exec.Command(executable, "daemon")
	cmd.Stdout = logFile
	cmd.Stderr = logFile
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	fmt.Printf("Daemon started with PID: %d\n", cmd.Process.Pid)
	return nil
}