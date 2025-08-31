package main

import (
	"fmt"
	"log"
	"net"

	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/SomeHowMicroservice/shm-be/auth/consumers"
	"github.com/SomeHowMicroservice/shm-be/auth/initialization"
	"github.com/SomeHowMicroservice/shm-be/auth/server"
)

var (
	userAddr = "localhost:8082"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Tải cấu hình Auth Service thất bại: %v", err)
	}

	rdb, err := initialization.InitCache(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	defer rdb.Close()

	mqc, err := initialization.InitMessageQueue(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	defer mqc.Close()

	userAddr = fmt.Sprintf("%s:%d", cfg.App.ServerHost, cfg.Services.UserPort)
	clients, err := initialization.InitClients(userAddr)
	if err != nil {
		log.Fatalf("Kết nối tới các dịch vụ khác thất bại: %v", err)
	}
	defer clients.Close()

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.GRPCPort))
	if err != nil {
		log.Fatalf("Không thể lắng nghe: %v", err)
	}
	defer lis.Close()

	grpcServer := server.NewGRPCServer(cfg, rdb, mqc.Chann, clients.UserClient)

	go consumers.StartSendEmailConsumer(mqc, grpcServer.Mailer)

	log.Println("Khởi chạy service thành công")
	if err := grpcServer.Server.Serve(lis); err != nil {
		log.Fatalf("Chạy gRPC server thất bại: %v", err)
	}
}
