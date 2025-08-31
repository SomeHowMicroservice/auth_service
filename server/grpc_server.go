package server

import (
	"time"

	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/SomeHowMicroservice/shm-be/auth/container"
	authpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/auth"
	userpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/user"
	"github.com/SomeHowMicroservice/shm-be/auth/smtp"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type GRPCServer struct {
	Server *grpc.Server
	Mailer smtp.SMTPService
}

func NewGRPCServer(cfg *config.Config, rdb *redis.Client, mqChann *amqp091.Channel, userClient userpb.UserServiceClient) *GRPCServer {
	kaParams := keepalive.ServerParameters{
		Time:                  5 * time.Minute,
		Timeout:               20 * time.Second,
		MaxConnectionIdle:     0,
		MaxConnectionAge:      0,
		MaxConnectionAgeGrace: 0,
	}

	kaPolicy := keepalive.EnforcementPolicy{
		MinTime:             1 * time.Minute,
		PermitWithoutStream: true,
	}

	grpcServer := grpc.NewServer(
		grpc.KeepaliveParams(kaParams),
		grpc.KeepaliveEnforcementPolicy(kaPolicy),
	)

	authContainer := container.NewContainer(cfg, rdb, mqChann, grpcServer, userClient)

	authpb.RegisterAuthServiceServer(grpcServer, authContainer.GRPCHandler)

	return &GRPCServer{
		grpcServer,
		authContainer.SMTPService,
	}
}
