package main

import (
	"fmt"
	"log"
	"net"

	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/SomeHowMicroservice/shm-be/auth/consumers"
	"github.com/SomeHowMicroservice/shm-be/auth/container"
	"github.com/SomeHowMicroservice/shm-be/auth/initialization"
	authpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/auth"
	"google.golang.org/grpc"
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

	mqc, err := initialization.InitMessageQueue(cfg)
	if err != nil {
		log.Fatalln(err)
	}
	defer mqc.Close()

	userAddr = cfg.App.ServerHost + fmt.Sprintf(":%d", cfg.Services.UserPort)
	clients := initialization.InitClients(userAddr)

	grpcServer := grpc.NewServer()
	authContainer := container.NewContainer(cfg, rdb, mqc.Chann, grpcServer, clients.UserClient)
	authpb.RegisterAuthServiceServer(grpcServer, authContainer.GRPCHandler)

	go consumers.StartSendEmailConsumer(mqc, authContainer.SMTPService)

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.GRPCPort))
	if err != nil {
		log.Fatalf("Không thể lắng nghe: %v", err)
	}
	defer lis.Close()

	log.Println("Khởi chạy service thành công")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Kết nối tới phục vụ thất bại: %v", err)
	}
}
