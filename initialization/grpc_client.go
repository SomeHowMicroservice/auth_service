package initialization

import (
	"fmt"
	"time"

	userpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GRPCClients struct {
	UserClient userpb.UserServiceClient
	userConn   *grpc.ClientConn
}

func InitClients(userAddr string) (*GRPCClients, error) {
	opts := dialOptions()
	userConn, err := grpc.NewClient(userAddr, opts...)
	if err != nil {
		return nil, fmt.Errorf("không thể kết nối tới User Service: %w", err)
	}
	userClient := userpb.NewUserServiceClient(userConn)

	return &GRPCClients{userClient, userConn}, nil
}

func (g *GRPCClients) Close() {
	_ = g.userConn.Close()
}

func dialOptions() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultServiceConfig(`{
			"methodConfig": [{
				"name": [{}],
				"retryPolicy": {
					"MaxAttempts": 4,
					"InitialBackoff": "0.1s",
					"MaxBackoff": "1s", 
					"BackoffMultiplier": 2.0,
					"RetryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
				}
			}]
		}`),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                5 * time.Minute,
			Timeout:             10 * time.Second,
			PermitWithoutStream: true,
		}),
	}
}
