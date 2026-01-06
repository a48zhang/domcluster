package main

import (
	"context"
	"fmt"
	"os"
	"os/user"
	"path/filepath"

	"domclusterd/cli"
	"domclusterd/daemon"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const defaultNodeID = "node-001"
const defaultNodeName = "Worker Node 1"

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "daemon":
		nodeID := defaultNodeID
		nodeName := defaultNodeName
		if len(os.Args) > 2 {
			nodeID = os.Args[2]
		}
		if len(os.Args) > 3 {
			nodeName = os.Args[3]
		}
		runDaemon(nodeID, nodeName)
	case "start":
		nodeID := defaultNodeID
		nodeName := defaultNodeName
		if len(os.Args) > 2 {
			nodeID = os.Args[2]
		}
		if len(os.Args) > 3 {
			nodeName = os.Args[3]
		}
		if err := cli.Start(nodeID, nodeName); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "stop":
		if err := cli.Stop(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "status":
		if err := cli.Status(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "logs":
		lines := 50 // 默认显示 50 行
		if len(os.Args) > 2 {
			fmt.Sscanf(os.Args[2], "%d", &lines)
		}
		if err := cli.Logs(lines); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "restart":
		if err := cli.Restart(); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runDaemon(nodeID, nodeName string) {
	// 检查是否为 root 用户
	currentUser, err := user.Current()
	if err != nil {
		panic(fmt.Sprintf("Failed to get current user: %v", err))
	}
	if currentUser.Uid != "0" {
		fmt.Fprintf(os.Stderr, "Error: domclusterd must be run as root user\n")
		fmt.Fprintf(os.Stderr, "Current user: %s (UID: %s)\n", currentUser.Username, currentUser.Uid)
		fmt.Fprintf(os.Stderr, "Please use: sudo domclusterd daemon\n")
		os.Exit(1)
	}

	// 确定日志文件目录
	logDir := "/var/log/domclusterd"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		// 如果没有权限，使用 ~/.local/log/domclusterd/
		homeDir, err := os.UserHomeDir()
		if err == nil {
			logDir = filepath.Join(homeDir, ".local", "log", "domclusterd")
			os.MkdirAll(logDir, 0755)
		} else {
			// 如果都失败，使用当前目录
			logDir = "."
		}
	}
	logFile := filepath.Join(logDir, "domclusterd.log")

	// 初始化 zap logger（输出到文件）
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(file),
		zapcore.InfoLevel,
	)

	logger := zap.New(core)
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	// 创建守护进程
	d, err := daemon.NewDaemon(nodeID, nodeName)
	if err != nil {
		zap.L().Fatal("Failed to create daemon", zap.Error(err))
	}

	// 运行守护进程
	ctx := context.Background()
	if err := d.Run(ctx, nodeID, nodeName); err != nil {
		zap.L().Fatal("Daemon error", zap.Error(err))
	}
}

func printUsage() {
	fmt.Println("Usage: domclusterd <command> [nodeID] [nodeName]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  daemon [nodeID] [nodeName]  Run as daemon process")
	fmt.Println("  start [nodeID] [nodeName]  Start daemon in background")
	fmt.Println("  stop                      Stop daemon")
	fmt.Println("  status                    Show daemon status")
	fmt.Println("  logs [n]                  Show last n lines of logs (default: 50)")
	fmt.Println("  restart                   Restart daemon")
}
