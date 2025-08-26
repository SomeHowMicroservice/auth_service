package initialization

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/redis/go-redis/v9"
)

func InitCache(cfg *config.Config) (*redis.Client, error) {
	rAddr := cfg.Cache.CHost + fmt.Sprintf(":%d", cfg.Cache.CPort)
	rdb := redis.NewClient(&redis.Options{
		Addr: rAddr,
		Password: cfg.Cache.CPassword,
		TLSConfig: &tls.Config{},
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("kết nối Redis thất bại: %v", err)
	}

	return rdb, nil
}