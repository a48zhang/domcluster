package main

import (
	"context"
	"fmt"
	"os"

	"d8rctl/cli"
	"d8rctl/config"
	"d8rctl/daemon"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "daemon":
		runDaemon()
	case "start":
		if err := cli.Start(); err != nil {
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
		lines := 50
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
	case "password":
		if err := cli.Password(os.Args[2:]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "pod":
		if len(os.Args) < 3 {
			fmt.Println("Usage: d8rctl pod <command>")
			fmt.Println("Commands:")
			fmt.Println("  list    List all connected domclusterd nodes")
			os.Exit(1)
		}
		podCommand := os.Args[2]
		switch podCommand {
		case "list":
			if err := cli.PodList(); err != nil {
				fmt.Printf("Error: %v\n", err)
				os.Exit(1)
			}
		default:
			fmt.Printf("Unknown pod command: %s\n", podCommand)
			os.Exit(1)
		}
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func runDaemon() {
	// 确保目录存在
	if err := config.EnsureDirs(); err != nil {
		fmt.Printf("Failed to create directories: %v\n", err)
		os.Exit(1)
	}

	logFile, err := os.OpenFile(config.GetLogFile(), os.O_CREATE|os.O_WRONLY|os.O_APPEND, daemon.FilePermission)
	if err != nil {
		fmt.Printf("Failed to open log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()

	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderCfg),
		zapcore.AddSync(logFile),
		zapcore.InfoLevel,
	)

	logger := zap.New(core)
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	d, err := daemon.NewDaemon()
	if err != nil {
		logger.Sync()
		zap.L().Fatal("Failed to create daemon", zap.Error(err))
	}

	ctx := context.Background()
	if err := d.Run(ctx); err != nil {
		logger.Sync()
		zap.L().Fatal("Daemon error", zap.Error(err))
	}
}

func printUsage() {
	fmt.Println("Usage: d8rctl <command>")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  daemon           Run as daemon process")
	fmt.Println("  start            Start daemon in background")
	fmt.Println("  stop             Stop daemon")
	fmt.Println("  status           Show daemon status")
	fmt.Println("  logs [n]         Show last n lines of logs (default: 50)")
	fmt.Println("  restart          Restart daemon")
	fmt.Println("  password [reset] Show password info or reset password")
	fmt.Println("  pod list         List all connected domclusterd nodes")
}