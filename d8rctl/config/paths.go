package config

import (
	"os"
	"path/filepath"
)

// GetPIDDir 获取PID文件目录
func GetPIDDir() string {
	return "/run/d8rctl"
	
}

// GetLogDir 获取日志文件目录
func GetLogDir() string {
	return "/var/log/d8rctl" 
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