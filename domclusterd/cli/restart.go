package cli

import (
	"fmt"

	"domclusterd/daemon"
)

// Restart 重启守护进程
func Restart() error {
	if !daemon.IsRunning() {
		return fmt.Errorf("daemon is not running")
	}

	if err := daemon.Restart(); err != nil {
		return fmt.Errorf("failed to restart daemon: %w", err)
	}

	fmt.Println("Daemon restarted")
	return nil
}