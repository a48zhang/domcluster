package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"d8rctl/connections"
	"d8rctl/services"
	pb "domcluster/api/proto"
	"go.uber.org/zap"
)

func main() {
	// 初始化 zap logger
	var logger *zap.Logger
	var err error
	if true { // debug mode
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	zap.ReplaceGlobals(logger)

	config := &connections.Config{
		Address:  ":50051",
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
	}

	server, err := connections.NewServer(config)
	if err != nil {
		zap.L().Fatal("Failed to create server", zap.Error(err))
	}

	// 注册 gRPC 服务
	pb.RegisterDomclusterServiceServer(server.GetServer(), services.NewDomclusterServer())
	zap.L().Sugar().Info("gRPC services registered")

	// 创建 context 用于优雅退出
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 监听中断信号
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan
		zap.L().Sugar().Info("Received shutdown signal")
		cancel()
	}()

	// 启动服务器
	zap.L().Sugar().Info("Server starting...")
	if err := server.Start(ctx); err != nil {
		zap.L().Fatal("Server error", zap.Error(err))
	}
}