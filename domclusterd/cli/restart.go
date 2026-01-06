package cli

import (
	"fmt"
	"os/user"

	"domclusterd/daemon"
)

// Restart 重启守护进程
func Restart() error {
	// 检查是否为 root 用户
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	if currentUser.Uid != "0" {
		return fmt.Errorf("domclusterd must be run as root user. Please use: sudo domclusterd restart")
	}

	if !daemon.IsRunning() {
		return fmt.Errorf("daemon is not running")
	}

	if err := daemon.Restart(); err != nil {
		return fmt.Errorf("failed to restart daemon: %w", err)
	}

	fmt.Println("Daemon restarted")
	return nil
}