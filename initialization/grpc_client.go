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
	userConn   *grpc.ClientConn
}

func InitClients(userAddr string) (*GRPCClients, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	userConn, err := grpc.DialContext(ctx, userAddr, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới User Service: %w", err)
	}
	userClient := userpb.NewUserServiceClient(userConn)

	return &GRPCClients{userClient, userConn}, nil
}

func (g *GRPCClients) Close() {
	_ = g.userConn.Close()
}
