package config

import (
	"os"
	"path/filepath"
)

// GetPIDDir 获取PID文件目录
func GetPIDDir() string {
	// 使用全局位置 /run/d8rctl，确保所有用户都能访问同一个守护进程实例
	pidDir := "/run/d8rctl"
	if err := os.MkdirAll(pidDir, 0755); err == nil {
		return pidDir
	}

	// 回退到 /var/run/d8rctl
	pidDir = "/var/run/d8rctl"
	if err := os.MkdirAll(pidDir, 0755); err == nil {
		return pidDir
	}

	// 最后回退到 /tmp
	return "/tmp"
}

// GetLogDir 获取日志文件目录
func GetLogDir() string {
	// 使用全局位置 /var/log/d8rctl，确保所有用户都能访问同一个日志文件
	logDir := "/var/log/d8rctl"
	if err := os.MkdirAll(logDir, 0755); err == nil {
		return logDir
	}

	// 回退到 /tmp
	return "/tmp"
}

// GetPIDFile 获取PID文件路径
func GetPIDFile() string {
	return filepath.Join(GetPIDDir(), "d8rctl.pid")
}

// GetLogFile 获取日志文件路径
func GetLogFile() string {
	return filepath.Join(GetLogDir(), "d8rctl.log")
}

// EnsureDirs 确保所有必要的目录存在
func EnsureDirs() error {
	// 确保PID目录存在
	if err := os.MkdirAll(GetPIDDir(), 0755); err != nil {
		return err
	}

	// 确保日志目录存在
	if err := os.MkdirAll(GetLogDir(), 0755); err != nil {
		return err
	}

	return nil
}

// GetSessionFile 获取会话令牌文件路径
func GetSessionFile() string {
	return filepath.Join("/run/d8rctl", "session")
}