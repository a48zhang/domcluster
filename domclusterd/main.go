package main

import (
	"context"
	"domclusterd/config"
	"domclusterd/connections"
	"domclusterd/monitor"
	"domclusterd/tasks"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}

	// 初始化 zap logger
	var logger *zap.Logger
	if false { // debug mode
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	zap.L().Info("Starting domclusterd",
		zap.String("address", cfg.GetAddress()),
		zap.Bool("tls", cfg.GetUseTLS()),
	)

	manager := connections.NewManager(&connections.Config{
		Address:  cfg.GetAddress(),
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		Timeout:  cfg.GetTimeout(),
	})

	ctx := context.Background()

	if err := manager.Start(ctx, "node-001", "Worker Node 1"); err != nil {
		zap.L().Fatal("Failed to start connection manager", zap.Error(err))
	}

	// 创建监控器
	m := monitor.NewMonitor(ctx)

	// 创建并启动状态报告器（定时上报）
	reporter := monitor.NewStatusReporter(m, manager)
	go reporter.Start(2 * time.Second) // 每30秒上报一次状态
	defer reporter.Stop()

	// 创建并注册查询处理器
	queryHandler := monitor.NewQueryHandler(m, manager)
	queryHandler.Register()
	defer queryHandler.Stop()

	tm := tasks.NewTaskManager(ctx)
	tm.Run()
	defer tm.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	zap.L().Info("Shutting down...")
	manager.Close()
}
