package cli

import (
	"encoding/json"
	"fmt"

	"d8rctl/daemon"
)

// Status 查看守护进程状态
func Status() error {
	if !daemon.IsRunning() {
		fmt.Println("Daemon is not running")
		return nil
	}

	status, err := daemon.GetStatus()
	if err != nil {
		return fmt.Errorf("failed to get status: %w", err)
	}

	data, _ := json.MarshalIndent(status, "", "  ")
	fmt.Println(string(data))

	return nil
}

// IsRunning 检查守护进程是否在运行
func IsRunning() bool {
	return daemon.IsRunning()
}