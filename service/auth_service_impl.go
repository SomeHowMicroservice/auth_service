package service

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/SomeHowMicroservice/shm-be/auth/common"
	"github.com/SomeHowMicroservice/shm-be/auth/config"
	"github.com/SomeHowMicroservice/shm-be/auth/mq"
	authpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/auth"
	userpb "github.com/SomeHowMicroservice/shm-be/auth/protobuf/user"
	"github.com/SomeHowMicroservice/shm-be/auth/repository"
	"github.com/SomeHowMicroservice/shm-be/auth/security"
	"github.com/SomeHowMicroservice/shm-be/auth/smtp"
	"github.com/google/uuid"
	"github.com/rabbitmq/amqp091-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type authServiceImpl struct {
	cacheRepo  repository.CacheRepository
	userClient userpb.UserServiceClient
	mailer     smtp.SMTPService
	cfg        *config.Config
	mqChannel  *amqp091.Channel
}

func NewAuthService(cacheRepo repository.CacheRepository, userClient userpb.UserServiceClient, mailer smtp.SMTPService, cfg *config.Config, mqChannel *amqp091.Channel) AuthService {
	return &authServiceImpl{
		cacheRepo,
		userClient,
		mailer,
		cfg,
		mqChannel,
	}
}

func (s *authServiceImpl) SignUp(ctx context.Context, req *authpb.SignUpRequest) (string, error) {
	uRes, err := s.userClient.CheckUsernameExists(ctx, &userpb.CheckUsernameExistsRequest{Username: req.Username})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return "", fmt.Errorf("lỗi từ user service: %s", st.Message())
		}
		return "", fmt.Errorf("lỗi không xác định: %w", err)
	}
	if uRes.Exists {
		return "", common.ErrUsernameAlreadyExists
	}

	eRes, err := s.userClient.CheckEmailExists(ctx, &userpb.CheckEmailExistsRequest{Email: req.Email})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return "", fmt.Errorf("lỗi từ user service: %s", st.Message())
		}
		return "", fmt.Errorf("lỗi không xác định: %w", err)
	}
	if eRes.Exists {
		return "", common.ErrEmailAlreadyExists
	}

	otp := common.GenerateOTP(6)
	registrationToken := uuid.NewString()

	hashedPassword, err := common.HashPassword(req.Password)
	if err != nil {
		return "", fmt.Errorf("băm mật khẩu thất bại: %w", err)
	}

	registrationData := &common.RegistrationData{
		Email:    req.Email,
		Username: req.Username,
		Password: hashedPassword,
		Otp:      otp,
		Attempts: 0,
	}

	if err = s.cacheRepo.SaveRegistrationData(ctx, registrationToken, registrationData, 3*time.Minute); err != nil {
		return "", err
	}

	emailMsg := common.AuthEmailMessage{
		To:      req.Email,
		Subject: "Xác thực đăng ký tài khoản tại SomeHow",
		Otp:     otp,
	}

	body, err := json.Marshal(emailMsg)
	if err != nil {
		return "", fmt.Errorf("chuyển đổi EmailMessage thất bại: %w", err)
	}

	if err := mq.PublishMessage(s.mqChannel, common.Exchange, common.RoutingKey, body); err != nil {
		return "", fmt.Errorf("publish email msg thất bại: %w", err)
	}

	return registrationToken, nil
}

func (s *authServiceImpl) VerifySignUp(ctx context.Context, req *authpb.VerifySignUpRequest) (*authpb.LoggedInResponse, error) {
	regData, err := s.cacheRepo.GetRegistrationData(ctx, req.RegistrationToken)
	if err != nil {
		return nil, err
	}
	if regData == nil {
		return nil, common.ErrAuthDataNotFound
	}

	if regData.Attempts >= 3 {
		if err = s.cacheRepo.DeleteAuthData(ctx, "sign-up", req.RegistrationToken); err != nil {
			return nil, err
		}
		return nil, common.ErrTooManyAttempts
	}

	regData.Attempts++
	if err = s.cacheRepo.SaveRegistrationData(ctx, req.RegistrationToken, regData, 3*time.Minute); err != nil {
		return nil, err
	}

	if regData.Otp != req.Otp {
		return nil, common.ErrInvalidOTP
	}

	uRes, err := s.userClient.CheckUsernameExists(ctx, &userpb.CheckUsernameExistsRequest{Username: regData.Username})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}
	if uRes.Exists {
		return nil, common.ErrUsernameAlreadyExists
	}

	eRes, err := s.userClient.CheckEmailExists(ctx, &userpb.CheckEmailExistsRequest{Email: regData.Email})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}
	if eRes.Exists {
		return nil, common.ErrEmailAlreadyExists
	}

	newUser := &userpb.CreateUserRequest{
		Username: regData.Username,
		Email:    regData.Email,
		Password: regData.Password,
	}
	userRes, err := s.userClient.CreateUser(ctx, newUser)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	accessToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.AccessExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("tạo token xác thực thất bại")
	}
	refreshToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.RefreshExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("tạo token xác thực thất bại")
	}

	if err = s.cacheRepo.DeleteAuthData(ctx, "sign-up", req.RegistrationToken); err != nil {
		return nil, err
	}

	authRes := &authpb.AuthResponse{
		Id:        userRes.Id,
		Username:  userRes.Username,
		Email:     userRes.Email,
		CreatedAt: userRes.CreatedAt,
		Profile: &authpb.ProfileResponse{
			Id:        userRes.Profile.Id,
			FirstName: userRes.Profile.FirstName,
			LastName:  userRes.Profile.LastName,
			Gender:    userRes.Profile.Gender,
			Dob:       userRes.Profile.Dob,
		},
	}

	return toLoggedInResponse(authRes, accessToken, s.cfg.Jwt.AccessExpiresIn, refreshToken, s.cfg.Jwt.RefreshExpiresIn), nil
}

func (s *authServiceImpl) SignIn(ctx context.Context, req *authpb.SignInRequest) (*authpb.LoggedInResponse, error) {
	userRes, err := s.userClient.GetUserByUsername(ctx, &userpb.GetUserByUsernameRequest{Username: req.Username})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	isCorrectPassword := common.VerifyPassword(req.Password, userRes.Password)
	if !isCorrectPassword {
		return nil, common.ErrInvalidPassword
	}

	if !hasRoleUser("user", userRes.Roles) {
		return nil, common.ErrForbidden
	}

	accessToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.AccessExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("tạo token xác thực thất bại")
	}
	refreshToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.RefreshExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("tạo token xác thực thất bại")
	}
	authRes := &authpb.AuthResponse{
		Id:        userRes.Id,
		Username:  userRes.Username,
		Email:     userRes.Email,
		CreatedAt: userRes.CreatedAt,
		Profile: &authpb.ProfileResponse{
			Id:        userRes.Profile.Id,
			FirstName: userRes.Profile.FirstName,
			LastName:  userRes.Profile.LastName,
			Gender:    userRes.Profile.Gender,
			Dob:       userRes.Profile.Dob,
		},
	}
	return toLoggedInResponse(authRes, accessToken, s.cfg.Jwt.AccessExpiresIn, refreshToken, s.cfg.Jwt.RefreshExpiresIn), nil
}

func (s *authServiceImpl) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (string, time.Duration, string, time.Duration, error) {
	accessToken, err := security.GenerateToken(req.Id, req.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.AccessExpiresIn)
	if err != nil {
		return "", 0, "", 0, fmt.Errorf("tạo token xác thực thất bại")
	}

	refreshToken, err := security.GenerateToken(req.Id, req.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.RefreshExpiresIn)
	if err != nil {
		return "", 0, "", 0, fmt.Errorf("tạo token xác thực thất bại")
	}

	return accessToken, s.cfg.Jwt.AccessExpiresIn, refreshToken, s.cfg.Jwt.RefreshExpiresIn, nil
}

func (s *authServiceImpl) ChangePassword(ctx context.Context, req *authpb.ChangePasswordRequest) (string, time.Duration, string, time.Duration, error) {
	userRes, err := s.userClient.GetUserById(ctx, &userpb.GetUserByIdRequest{Id: req.Id})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return "", 0, "", 0, common.ErrUserNotFound
			default:
				return "", 0, "", 0, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return "", 0, "", 0, fmt.Errorf("lỗi không xác định: %w", err)
	}

	isCorrectPassword := common.VerifyPassword(req.OldPassword, userRes.Password)
	if !isCorrectPassword {
		return "", 0, "", 0, common.ErrInvalidPassword
	}

	hashedNewPassword, err := common.HashPassword(req.NewPassword)
	if err != nil {
		return "", 0, "", 0, fmt.Errorf("băm mật khẩu thất bại: %w", err)
	}

	_, err = s.userClient.UpdateUserPassword(ctx, &userpb.UpdateUserPasswordRequest{
		Id:          userRes.Id,
		NewPassword: hashedNewPassword,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return "", 0, "", 0, common.ErrUserNotFound
			default:
				return "", 0, "", 0, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return "", 0, "", 0, fmt.Errorf("lỗi không xác định: %w", err)
	}
	accessToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.AccessExpiresIn)
	if err != nil {
		return "", 0, "", 0, fmt.Errorf("tạo token xác thực thất bại")
	}
	refreshToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.RefreshExpiresIn)
	if err != nil {
		return "", 0, "", 0, fmt.Errorf("tạo token xác thực thất bại")
	}
	return accessToken, s.cfg.Jwt.AccessExpiresIn, refreshToken, s.cfg.Jwt.RefreshExpiresIn, nil
}

func (s *authServiceImpl) ForgotPassword(ctx context.Context, req *authpb.ForgotPasswordRequest) (string, error) {
	res, err := s.userClient.CheckEmailExists(ctx, &userpb.CheckEmailExistsRequest{
		Email: req.Email,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			return "", fmt.Errorf("lỗi từ user service: %s", st.Message())
		}
		return "", fmt.Errorf("lỗi không xác định: %w", err)
	}
	if !res.Exists {
		return "", common.ErrUserNotFound
	}

	otp := common.GenerateOTP(6)
	forgotPasswordToken := uuid.NewString()

	forgotPasswordData := &common.ForgotPasswordData{
		Email:    req.Email,
		Otp:      otp,
		Attempts: 0,
	}

	if err = s.cacheRepo.SaveForgotPasswordData(ctx, forgotPasswordToken, forgotPasswordData, 3*time.Minute); err != nil {
		return "", err
	}

	emailMsg := common.AuthEmailMessage{
		To:      req.Email,
		Subject: "Xác thực quên mật khẩu tại SomeHow",
		Otp:     otp,
	}

	body, err := json.Marshal(emailMsg)
	if err != nil {
		return "", fmt.Errorf("chuyển đổi EmailMessage thất bại: %w", err)
	}

	if err := mq.PublishMessage(s.mqChannel, common.Exchange, common.RoutingKey, body); err != nil {
		return "", fmt.Errorf("publish email msg thất bại: %w", err)
	}

	return forgotPasswordToken, nil
}

func (s *authServiceImpl) VerifyForgotPassword(ctx context.Context, req *authpb.VerifyForgotPasswordRequest) (string, error) {
	forgData, err := s.cacheRepo.GetForgotPasswordData(ctx, req.ForgotPasswordToken)
	if err != nil {
		return "", err
	}
	if forgData == nil {
		return "", common.ErrAuthDataNotFound
	}

	if forgData.Attempts >= 3 {
		if err = s.cacheRepo.DeleteAuthData(ctx, "forgot-password", req.ForgotPasswordToken); err != nil {
			return "", err
		}
		return "", common.ErrTooManyAttempts
	}

	if forgData.Otp != req.Otp {
		return "", common.ErrInvalidOTP
	}

	resetPasswordToken := uuid.NewString()

	if err = s.cacheRepo.SaveResetPasswordData(ctx, resetPasswordToken, forgData.Email, 3*time.Minute); err != nil {
		return "", err
	}

	if err = s.cacheRepo.DeleteAuthData(ctx, "forgot-password", req.ForgotPasswordToken); err != nil {
		return "", err
	}

	return resetPasswordToken, nil
}

func (s *authServiceImpl) ResetPassword(ctx context.Context, req *authpb.ResetPasswordRequest) error {
	email, err := s.cacheRepo.GetResetPasswordData(ctx, req.ResetPasswordToken)
	if err != nil {
		return err
	}
	if email == "" {
		return common.ErrAuthDataNotFound
	}

	userRes, err := s.userClient.GetUserPublicByEmail(ctx, &userpb.GetUserByEmailRequest{
		Email: email,
	})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return common.ErrUserNotFound
			default:
				return fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return fmt.Errorf("lỗi không xác định: %w", err)
	}

	hashedNewPassword, err := common.HashPassword(req.NewPassword)
	if err != nil {
		return fmt.Errorf("băm mật khẩu thất bại: %w", err)
	}

	if _, err = s.userClient.UpdateUserPassword(ctx, &userpb.UpdateUserPasswordRequest{
		Id:          userRes.Id,
		NewPassword: hashedNewPassword,
	}); err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return common.ErrUserNotFound
			default:
				return fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return fmt.Errorf("lỗi không xác định: %w", err)
	}

	if err = s.cacheRepo.DeleteAuthData(ctx, "reset-password", req.ResetPasswordToken); err != nil {
		return err
	}

	return nil
}

func (s *authServiceImpl) AdminSignIn(ctx context.Context, req *authpb.SignInRequest) (*authpb.LoggedInResponse, error) {
	userRes, err := s.userClient.GetUserByUsername(ctx, &userpb.GetUserByUsernameRequest{Username: req.Username})
	if err != nil {
		if st, ok := status.FromError(err); ok {
			switch st.Code() {
			case codes.NotFound:
				return nil, common.ErrUserNotFound
			default:
				return nil, fmt.Errorf("lỗi từ user service: %s", st.Message())
			}
		}
		return nil, fmt.Errorf("lỗi không xác định: %w", err)
	}

	isCorrectPassword := common.VerifyPassword(req.Password, userRes.Password)
	if !isCorrectPassword {
		return nil, common.ErrInvalidPassword
	}

	if !isAdmin("user", userRes.Roles) {
		return nil, common.ErrForbidden
	}

	accessToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.AccessExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("tạo token xác thực thất bại")
	}
	refreshToken, err := security.GenerateToken(userRes.Id, userRes.Roles, s.cfg.Jwt.SecretKey, s.cfg.Jwt.RefreshExpiresIn)
	if err != nil {
		return nil, fmt.Errorf("tạo token xác thực thất bại")
	}
	authRes := &authpb.AuthResponse{
		Id:        userRes.Id,
		Username:  userRes.Username,
		Email:     userRes.Email,
		CreatedAt: userRes.CreatedAt,
		Profile: &authpb.ProfileResponse{
			Id:        userRes.Profile.Id,
			FirstName: userRes.Profile.FirstName,
			LastName:  userRes.Profile.LastName,
			Gender:    userRes.Profile.Gender,
			Dob:       userRes.Profile.Dob,
		},
	}

	return toLoggedInResponse(authRes, accessToken, s.cfg.Jwt.AccessExpiresIn, refreshToken, s.cfg.Jwt.RefreshExpiresIn), nil
}

func toLoggedInResponse(user *authpb.AuthResponse, accessToken string, accessExpiresIn time.Duration, refreshToken string, refreshExpiresIn time.Duration) *authpb.LoggedInResponse {
	return &authpb.LoggedInResponse{
		User:             user,
		AccessToken:      accessToken,
		AccessExpiresIn:  int64(accessExpiresIn),
		RefreshToken:     refreshToken,
		RefreshExpiresIn: int64(refreshExpiresIn),
	}
}

func hasRoleUser(role string, roles []string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}

	return false
}

func isAdmin(role string, roles []string) bool {
	if len(roles) > 1 && hasRoleUser(role, roles) {
		return true
	}

	return false
}
