package main

import (
	"context"
	"domclusterd/tasks"
	"os"
	"os/signal"
	"syscall"

	"domclusterd/config"
	"domclusterd/connections"
	"domclusterd/log"

	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic(err)
	}
	if err := log.Init(false); err != nil {
		panic(err)
	}

	log.Info("Starting domclusterd",
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
		log.Fatal("Failed to start connection manager", zap.Error(err))
	}

	tm := tasks.NewTaskManager(ctx)
	tm.Run()
	defer tm.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down...")
	manager.Close()
}
