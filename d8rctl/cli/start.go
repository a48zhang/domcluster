package cli

import (
	"d8rctl/config"
	"fmt"
	"os"
	"os/exec"
	"syscall"
	"time"

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

	// 确保目录存在
	if err := config.EnsureDirs(); err != nil {
		return fmt.Errorf("failed to create directories: %w\n\nSuggestion: Try running with sudo to create system directories", err)
	}

	executable, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable: %w", err)
	}

	logFile, err := os.OpenFile(config.GetLogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file %s: %w\n\nSuggestion: Check directory permissions or run with sudo", config.GetLogFile(), err)
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

	pid := cmd.Process.Pid
	fmt.Printf("Daemon started with PID: %d\n", pid)
	fmt.Printf("Log file: %s\n", config.GetLogFile())
	fmt.Printf("PID file: %s\n", config.GetPIDFile())

	// 等待守护进程启动并验证
	for i := 0; i < 10; i++ {
		time.Sleep(100 * time.Millisecond)
		if IsRunning() {
			fmt.Println("Daemon is running successfully")
			return nil
		}
	}

	// 检查进程是否已退出
	if cmd.ProcessState != nil && cmd.ProcessState.Exited() {
		exitCode := cmd.ProcessState.ExitCode()
		return fmt.Errorf("daemon exited immediately with code %d\n\nTroubleshooting:\n1. Check logs: %s\n2. Verify permissions for directories:\n   - %s\n   - %s\n3. Try running with sudo\n4. Check if another instance is running",
			exitCode, config.GetLogFile(), config.GetPIDDir(), config.GetLogDir())
	}

	return fmt.Errorf("daemon started but is not responding\n\nTroubleshooting:\n1. Check logs: %s\n2. Verify process status: ps -p %d\n3. Check socket file: /run/d8rctl/d8rctl.sock",
		config.GetLogFile(), pid)
}