package initialization

import (
	"context"
	"fmt"
	"time"

	userpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type GRPCClients struct {
	UserClient userpb.UserServiceClient
}

func InitClients(userAddr string) *GRPCClients {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userConn, err := grpc.DialContext(ctx, userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		panic(fmt.Errorf("kết nối tới User Service thất bại: %w", err))
	}
	userClient := userpb.NewUserServiceClient(userConn)

	return &GRPCClients{userClient}
}
