package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/SomeHowMicroservice/shm-be/auth/initialization"
	"github.com/SomeHowMicroservice/shm-be/auth/mq"
	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/ThreeDotsLabs/watermill/message/router/middleware"
	"github.com/redis/go-redis/v9"
)

var (
	userAddr = "localhost:8082"
)

type Server struct {
	grpcServer *GRPCServer
	lis        net.Listener
	rdb        *redis.Client
	clients    *initialization.GRPCClients
	router     *message.Router
	watermill  *initialization.WatermillConnection
}

func NewServer(cfg *config.Config) (*Server, error) {
	rdb, err := initialization.InitCache(cfg)
	if err != nil {
		return nil, err
	}

	userAddr = fmt.Sprintf("%s:%d", cfg.App.ServerHost, cfg.Services.UserPort)
	clients, err := initialization.InitClients(userAddr)
	if err != nil {
		return nil, err
	}

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", cfg.App.GRPCPort))
	if err != nil {
		return nil, err
	}

	logger := watermill.NewStdLogger(false, false)
	router, err := message.NewRouter(message.RouterConfig{}, logger)
	if err != nil {
		return nil, err
	}

	router.AddMiddleware(
		middleware.CorrelationID,
		middleware.Retry{
			MaxRetries:      5,
			InitialInterval: time.Second,
			Multiplier:      2.0,
			MaxInterval:     10 * time.Second,
			Logger:          logger,
		}.Middleware,
		middleware.Recoverer,
	)

	wm, err := initialization.InitWatermill(cfg, logger)
	if err != nil {
		return nil, err
	}

	grpcServer := NewGRPCServer(cfg, rdb, wm.Publisher, clients.UserClient)

	mq.RegisterSendEmailConsumer(router, wm.Subscriber, grpcServer.Mailer)

	go func() {
		if err := router.Run(context.Background()); err != nil {
			log.Printf("Lỗi chạy message router: %v", err)
		}
	}()

	return &Server{
		grpcServer,
		lis,
		rdb,
		clients,
		router,
		wm,
	}, nil
}

func (s *Server) Start() error {
	return s.grpcServer.Server.Serve(s.lis)
}

func (s *Server) Shutdown(ctx context.Context) {
	log.Println("Đang shutdown service...")

	if s.router != nil {
		s.router.Close()
	}
	if s.watermill != nil {
		s.watermill.Close()
	}
	if s.grpcServer != nil {
		stopped := make(chan struct{})
		go func() {
			s.grpcServer.Server.GracefulStop()
			close(stopped)
		}()

		select {
		case <-ctx.Done():
			log.Println("Timeout khi dừng gRPC server, force stop...")
			s.grpcServer.Server.Stop()
		case <-stopped:
			log.Println("Đã shutdown gRPC server")
		}
	}
	if s.lis != nil {
		s.lis.Close()
	}
	if s.rdb != nil {
		s.rdb.Close()
	}
	if s.clients != nil {
		s.clients.Close()
	}

	log.Println("Shutdown service thành công")
}
