package container

import (
	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/SomeHowMicroservice/shm-be/auth/handler"
	"github.com/SomeHowMicroservice/shm-be/auth/repository"
	"github.com/SomeHowMicroservice/shm-be/auth/service"
	"github.com/SomeHowMicroservice/shm-be/auth/smtp"
	userpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/user"
	"github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
	"google.golang.org/grpc"
)

type Container struct {
	GRPCHandler *handler.GRPCHandler
	SMTPService smtp.SMTPService
}

func NewContainer(cfg *config.Config, rdb *redis.Client, mqChan *amqp091.Channel, grpcServer *grpc.Server, userClient userpb.UserServiceClient) *Container {
	mailer := smtp.NewSMTPService(cfg)
	cacheRepo := repository.NewCacheRepository(rdb)
	svc := service.NewAuthService(cacheRepo, userClient, mailer, cfg, mqChan)
	hdl := handler.NewGRPCHandler(grpcServer, svc)
	return &Container{hdl, mailer}
}
