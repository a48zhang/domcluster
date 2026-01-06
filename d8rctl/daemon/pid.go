package daemon

import (
	"d8rctl/config"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"go.uber.org/zap"
)

const (
	FilePermission = 0644 // 文件权限：所有者读写，组和其他用户只读
)

// getPIDFile 获取PID文件路径
func getPIDFile() string {
	return config.GetPIDFile()
}

// WritePID 写入 PID 文件
func WritePID(pid int) error {
	data := []byte(strconv.Itoa(pid))
	return os.WriteFile(getPIDFile(), data, FilePermission)
}

// ReadPID 读取 PID 文件
func ReadPID() (int, error) {
	data, err := os.ReadFile(getPIDFile())
	if err != nil {
		return 0, err
	}
	return strconv.Atoi(string(data))
}

// RemovePID 删除 PID 文件
func RemovePID() error {
	return os.Remove(getPIDFile())
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

	err = process.Signal(syscall.Signal(0))
	if err != nil {
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