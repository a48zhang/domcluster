package config

import (
	"os"
	"path/filepath"
)

// GetPIDDir 获取PID文件目录
func GetPIDDir() string {
	// 优先使用 XDG_RUNTIME_DIR (通常为 /run/user/<uid>)
	if runtimeDir := os.Getenv("XDG_RUNTIME_DIR"); runtimeDir != "" {
		if _, err := os.Stat(runtimeDir); err == nil {
			return runtimeDir
		}
	}

	// 尝试 /run (需要root权限)
	pidFile := filepath.Join("/run", "d8rctl.pid")
	file, err := os.OpenFile(pidFile, os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		file.Close()
		os.Remove(pidFile)
		return "/run"
	}

	// 最后回退到 /tmp
	return "/tmp"
}

// GetLogDir 获取日志文件目录
func GetLogDir() string {
	// 优先使用 XDG_STATE_HOME (通常为 ~/.local/state)
	if stateDir := os.Getenv("XDG_STATE_HOME"); stateDir != "" {
		return stateDir
	}

	// 回退到 ~/.local/state
	homeDir, err := os.UserHomeDir()
	if err == nil {
		stateDir := filepath.Join(homeDir, ".local", "state")
		// 尝试创建目录
		os.MkdirAll(stateDir, 0755)
		if _, err := os.Stat(stateDir); err == nil {
			return stateDir
		}
	}

	// 尝试 /var/log (需要root权限)
	if _, err := os.Stat("/var/log"); err == nil {
		if _, err := os.Stat(filepath.Join("/var/log", "d8rctl.log")); err == nil || os.IsPermission(err) {
			return "/var/log"
		}
	}

	// 最后回退到 /tmp
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
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join("/tmp", ".d8rctl_session")
	}
	return filepath.Join(homeDir, ".d8rctl_session")
}