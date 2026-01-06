package cli

import (
	"fmt"
	"os"
	"syscall"

	"domclusterd/daemon"
)

// Status 查看守护进程状态
func Status() error {
	if !daemon.IsRunning() {
		fmt.Println("Daemon is not running")
		return nil
	}

	// 读取 PID 文件
	pid, err := daemon.ReadPID()
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	// 获取进程信息
	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	// 检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		fmt.Println("Daemon is not running (process not found)")
		return nil
	}

	// 简单输出状态信息
	uptime, err := getProcessUptime(pid)
	if err != nil {
		uptime = "unknown"
	}

	fmt.Printf("Daemon is running\n")
	fmt.Printf("  PID: %d\n", pid)
	fmt.Printf("  Uptime: %s\n", uptime)

	return nil
}

// IsRunning 检查守护进程是否在运行
func IsRunning() bool {
	return daemon.IsRunning()
}

// getProcessUptime 获取进程运行时间（简化版本）
func getProcessUptime(pid int) (string, error) {
	// 在 Windows 上获取进程运行时间比较复杂
	// 这里返回一个简化版本
	return "unknown", nil
}