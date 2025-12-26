package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"d8rctl/connections"
	"d8rctl/services"
	pb "domcluster/api/proto"
)

func main() {
	config := &connections.Config{
		Address:  ":50051",
		CertFile: "",
		KeyFile:  "",
		CAFile:   "",
	}

	server, err := connections.NewServer(config)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	// 注册 gRPC 服务
	pb.RegisterDomclusterServiceServer(server.GetServer(), services.NewDomclusterServer())
	log.Println("gRPC services registered")

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		if err := server.Start(); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	<-quit
	log.Println("Shutting down server...")
	server.Stop()
}