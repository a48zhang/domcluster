package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"domclusterd/connections"
)

func main() {
	manager := connections.NewManager(&connections.Config{
		Address:  "localhost:50051",
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
		Timeout:  10 * time.Second,
	})

	err := manager.Connect()
	if err != nil {
		panic(err)
	}

	// 注册节点
	err = manager.RegisterNode("node-001", "Worker Node 1", "judgehost")
	if err != nil {
		panic(err)
	}

	fmt.Println("Node registered")

	// 心跳循环
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		for {
			select {
			case <-ticker.C:
				if err := manager.SendHeartbeat(); err != nil {
					fmt.Printf("Heartbeat failed: %v\n", err)
				}
			case <-quit:
				return
			}
		}
	}()

	<-quit
	fmt.Println("Shutting down...")
	manager.Close()
}