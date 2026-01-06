package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"syscall"

	"go.uber.org/zap"
)

var pidFile string

func init() {
	// 强制使用 /var/run/domclusterd/，确保所有用户访问同一个实例
	pidDir := "/var/run/domclusterd"
	if err := os.MkdirAll(pidDir, 0755); err != nil {
		panic(fmt.Sprintf("Failed to create PID directory %s: %v. Please run as root.", pidDir, err))
	}
	pidFile = filepath.Join(pidDir, "domclusterd.pid")
}

// WritePID 写入 PID 文件
func WritePID(pid int) error {
	data := []byte(strconv.Itoa(pid))
	return os.WriteFile(pidFile, data, 0644)
}

// ReadPID 读取 PID 文件
func ReadPID() (int, error) {
	data, err := os.ReadFile(pidFile)
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// RemovePID 删除 PID 文件
func RemovePID() error {
	return os.Remove(pidFile)
}

// IsRunning 检查进程是否在运行
func IsRunning() bool {
	pid, err := ReadPID()
	if err != nil {
		return false
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return false
	}

	// 发送信号 0 检查进程是否存在
	err = process.Signal(syscall.Signal(0))
	if err != nil {
		// 进程不存在
		RemovePID()
		return false
	}

	return true
}

// Stop 停止守护进程
func Stop() error {
	pid, err := ReadPID()
	if err != nil {
		return fmt.Errorf("failed to read PID file: %w", err)
	}

	process, err := os.FindProcess(pid)
	if err != nil {
		return fmt.Errorf("failed to find process: %w", err)
	}

	zap.L().Sugar().Infof("Sending SIGTERM to process %d", pid)
	if err := process.Signal(syscall.SIGTERM); err != nil {
		return fmt.Errorf("failed to send signal: %w", err)
	}

	return nil
}