package main

import (
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
		zap.String("role", cfg.GetRole()),
		zap.Bool("tls", cfg.GetUseTLS()),
	)

	manager := connections.NewManager(&connections.Config{
		Address:  cfg.GetAddress(),
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		Timeout:  cfg.GetTimeout(),
	})

	if err := manager.Connect(); err != nil {
		log.Fatal("Failed to connect to server", zap.Error(err))
	} else {
		log.Info("Connected to server")
	}

	if err := manager.RegisterNode("node-001", "Worker Node 1", cfg.GetRole()); err != nil {
		log.Fatal("Failed to register node", zap.Error(err))
	}

	log.Info("Node registered successfully")

	tm := tasks.TaskManager{}

	defer tm.Stop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	log.Info("Shutting down...")
	manager.Close()
}
