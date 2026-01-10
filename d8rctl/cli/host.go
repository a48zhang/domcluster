package cli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"

	"d8rctl/daemon"
	"d8rctl/services"
)

// HostAdd 添加主机
func HostAdd(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: d8rctl host add <user@host[:port]> [--password <password> | --key-file <path>] [--d8rctl-address <address>]")
	}

	sshConnStr := args[0]
	password := ""
	keyFile := ""
	d8rctlAddress := "localhost:50051"

	// 解析参数
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--password":
			if i+1 < len(args) {
				password = args[i+1]
				i++
			}
		case "--key-file", "--key":
			if i+1 < len(args) {
				keyFile = args[i+1]
				i++
			}
		case "--d8rctl-address", "--address":
			if i+1 < len(args) {
				d8rctlAddress = args[i+1]
				i++
			}
		}
	}

	// 如果没有提供密码和密钥文件，提示用户输入密码
	if password == "" && keyFile == "" {
		fmt.Print("Enter SSH password: ")
		fmt.Scanln(&password)
	}

	// 检查daemon是否运行
	if !daemon.IsRunning() {
		return fmt.Errorf("daemon is not running, please start it first with 'd8rctl start'")
	}

	// 创建请求
	req := services.HostProvisionRequest{
		SSHConnectionString: sshConnStr,
		Password:            password,
		KeyFile:             keyFile,
		D8rctlAddress:       d8rctlAddress,
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	// 发送请求
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(_ context.Context, _, _ string) (net.Conn, error) {
				return net.Dial("unix", daemon.GetCLISocketPath())
			},
		},
	}

	resp, err := client.Post("http://unix/hosts/add", "application/json", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		var errResp map[string]interface{}
		if err := json.Unmarshal(body, &errResp); err == nil {
			if msg, ok := errResp["error"].(string); ok {
				return fmt.Errorf("failed to add host: %s", msg)
			}
		}
		return fmt.Errorf("failed to add host: status %d", resp.StatusCode)
	}

	// 解析结果
	var result services.HostProvisionResult
	if err := json.Unmarshal(body, &result); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// 显示结果
	if result.Success {
		fmt.Println("✓ Host added successfully!")
		fmt.Printf("  Node ID:  %s\n", result.NodeID)
		fmt.Printf("  Hostname: %s\n", result.Hostname)
		fmt.Printf("  OS:       %s\n", result.OS)
		fmt.Printf("  Arch:     %s\n", result.Arch)
		fmt.Println("\nThe host should now appear in the node list.")
		fmt.Println("Use 'd8rctl pod list' to verify.")
	} else {
		fmt.Printf("✗ Failed to add host: %s\n", result.Message)
		if result.Hostname != "" {
			fmt.Printf("  Hostname: %s\n", result.Hostname)
			fmt.Printf("  OS:       %s\n", result.OS)
			fmt.Printf("  Arch:     %s\n", result.Arch)
		}
		return fmt.Errorf("host provision failed")
	}

	return nil
}

// HostList 列出主机
func HostList() error {
	// 复用 pod list 功能
	return PodList()
}

// HostRemove 移除主机
func HostRemove(args []string) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: d8rctl host remove <node_id>")
	}

	nodeID := args[0]
	fmt.Printf("Removing host %s is not yet implemented.\n", nodeID)
	fmt.Println("You can manually stop domclusterd on the remote host.")
	
	return nil
}

// Host 主机管理命令入口
func Host(args []string) error {
	if len(args) < 1 {
		fmt.Println("Usage: d8rctl host <command>")
		fmt.Println()
		fmt.Println("Commands:")
		fmt.Println("  add <user@host[:port]>   Add a new host")
		fmt.Println("    Options:")
		fmt.Println("      --password <password>       SSH password")
		fmt.Println("      --key-file <path>           SSH private key file")
		fmt.Println("      --d8rctl-address <address>  D8rctl server address (default: localhost:50051)")
		fmt.Println()
		fmt.Println("  list                     List all hosts")
		fmt.Println("  remove <node_id>         Remove a host")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  d8rctl host add root@192.168.1.100 --password mypassword")
		fmt.Println("  d8rctl host add user@server.com:22 --key-file ~/.ssh/id_rsa")
		return nil
	}

	cmd := args[0]
	switch cmd {
	case "add":
		return HostAdd(args[1:])
	case "list":
		return HostList()
	case "remove", "rm":
		return HostRemove(args[1:])
	default:
		return fmt.Errorf("unknown command: %s", cmd)
	}
}
