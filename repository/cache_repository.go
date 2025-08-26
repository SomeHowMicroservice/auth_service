package repository

import (
	"context"
	"time"

	"github.com/SomeHowMicroservice/shm-be/auth/common"
)

type CacheRepository interface {
	SaveRegistrationData(ctx context.Context, registrationToken string, data *common.RegistrationData, ttl time.Duration) error

	SaveForgotPasswordData(ctx context.Context, forgotPasswordToken string, data *common.ForgotPasswordData, ttl time.Duration) error

	SaveResetPasswordData(ctx context.Context, resetPasswordToken, email string, ttl time.Duration) error

	GetRegistrationData(ctx context.Context, registrationToken string) (*common.RegistrationData, error)

	GetForgotPasswordData(ctx context.Context, forgotPasswordToken string) (*common.ForgotPasswordData, error)

	GetResetPasswordData(ctx context.Context, resetPasswordToken string) (string, error)

	DeleteAuthData(ctx context.Context, keyType string, token string) error
}