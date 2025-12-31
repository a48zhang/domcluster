package cli

import (
	"fmt"

	"domclusterd/daemon"
)

// Stop 停止守护进程
func Stop() error {
	if !daemon.IsRunning() {
		return fmt.Errorf("daemon is not running")
	}

	if err := daemon.CallStop(); err != nil {
		return fmt.Errorf("failed to stop daemon: %w", err)
	}

	fmt.Println("Daemon stopped")
	return nil
}