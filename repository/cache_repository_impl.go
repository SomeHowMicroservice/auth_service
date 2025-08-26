package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SomeHowMicroservice/shm-be/auth/common"
	"github.com/redis/go-redis/v9"
)

type cacheRepositoryImpl struct {
	rdb *redis.Client
}

func NewCacheRepository(rdb *redis.Client) CacheRepository {
	return &cacheRepositoryImpl{rdb}
}

const (
	serviceName = "auth-service"
)

func (r *cacheRepositoryImpl) SaveRegistrationData(ctx context.Context, registrationToken string, data *common.RegistrationData, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("chuyển đổi dữ liệu thất bại: %w", err)
	}

	redisKey := fmt.Sprintf("%s:sign-up:%s", serviceName, registrationToken)
	if err := r.rdb.Set(ctx, redisKey, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("lưu dữ liệu thất bại: %w", err)
	}

	return nil
}

func (r *cacheRepositoryImpl) SaveForgotPasswordData(ctx context.Context, forgotPasswordToken string, data *common.ForgotPasswordData, ttl time.Duration) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("chuyển đổi dữ liệu thất bại: %w", err)
	}

	redisKey := fmt.Sprintf("%s:forgot-password:%s", serviceName, forgotPasswordToken)
	if err := r.rdb.Set(ctx, redisKey, bytes, ttl).Err(); err != nil {
		return fmt.Errorf("lưu dữ liệu thất bại: %w", err)
	}

	return nil
}

func (r *cacheRepositoryImpl) SaveResetPasswordData(ctx context.Context, resetPasswordToken, email string, ttl time.Duration) error {
	redisKey := fmt.Sprintf("%s:reset-password:%s", serviceName, resetPasswordToken)
	if err := r.rdb.Set(ctx, redisKey, email, ttl).Err(); err != nil {
		return fmt.Errorf("lưu dữ liệu thất bại: %w", err)
	}

	return nil
}

func (r *cacheRepositoryImpl) GetRegistrationData(ctx context.Context, registrationToken string) (*common.RegistrationData, error) {
	redisKey := fmt.Sprintf("%s:sign-up:%s", serviceName, registrationToken)
	regDataJSON, err := r.rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("lấy dữ liệu từ bộ nhớ tạm thất bại: %w", err)
	}

	var regData common.RegistrationData
	if err = json.Unmarshal([]byte(regDataJSON), &regData); err != nil {
		return nil, fmt.Errorf("chuyển đổi dữ liệu thất bại: %w", err)
	}

	return &regData, nil
}

func (r *cacheRepositoryImpl) GetForgotPasswordData(ctx context.Context, forgotPasswordToken string) (*common.ForgotPasswordData, error) {
	redisKey := fmt.Sprintf("%s:forgot-password:%s", serviceName, forgotPasswordToken)
	forgDataJSON, err := r.rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return nil, nil
	} else if err != nil {
		return nil, fmt.Errorf("lấy dữ liệu từ bộ nhớ tạm thất bại: %w", err)
	}

	var forgData common.ForgotPasswordData
	if err = json.Unmarshal([]byte(forgDataJSON), &forgData); err != nil {
		return nil, fmt.Errorf("chuyển đổi dữ liệu thất bại: %w", err)
	}

	return &forgData, nil
}

func (r *cacheRepositoryImpl) GetResetPasswordData(ctx context.Context, resetPasswordToken string) (string, error) {
	redisKey := fmt.Sprintf("%s:reset-password:%s", serviceName, resetPasswordToken)
	email, err := r.rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("lấy dữ liệu từ bộ nhớ tạm thất bại: %w", err)
	}

	return email, nil
}

func (r *cacheRepositoryImpl) DeleteAuthData(ctx context.Context, keyType string, token string) error {
	redisKey := fmt.Sprintf("%s:%s:%s", serviceName, keyType, token)
	if err := r.rdb.Del(ctx, redisKey).Err(); err != nil {
		return fmt.Errorf("xóa dữ liệu thất bại: %w", err)
	}

	return nil
}
