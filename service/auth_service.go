package service

import (
	"context"
	"time"

	authpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/auth"
)

type AuthService interface {
	SignUp(ctx context.Context, req *authpb.SignUpRequest) (string, error)

	VerifySignUp(ctx context.Context, req *authpb.VerifySignUpRequest) (*authpb.LoggedInResponse, error)

	SignIn(ctx context.Context, req *authpb.SignInRequest) (*authpb.LoggedInResponse, error)

	RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (string, time.Duration, string, time.Duration, error)

	ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (string, time.Duration, string, time.Duration, error)

	ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (string, error)

	VerifyForgotPassword(ctx context.Context, req *authpb.VerifyForgotPasswordRequest) (string, error)

	ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) error

	AdminSignIn(ctx context.Context, req *authpb.SignInRequest) (*authpb.LoggedInResponse, error)
}
