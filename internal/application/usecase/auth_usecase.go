package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/mail"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/sirupsen/logrus"
	tm "time"
)

type AuthUseCase struct {
	repo  repository.UserRepository
	jwt   jwt.Service
	redis redis.Service
	mail  *mail.Service
	cfg   *config.Config
}

func NewAuthUseCase(repo repository.UserRepository, jwt jwt.Service, redis redis.Service, mail *mail.Service, cfg *config.Config) *AuthUseCase {
	return &AuthUseCase{
		repo:  repo,
		jwt:   jwt,
		redis: redis,
		mail:  mail,
		cfg:   cfg,
	}
}

func (a *AuthUseCase) Register(req *dto.RegisterRequest, ctx context.Context) (*entity.User, error) {
	existingUser, _ := a.repo.FindByEmail(ctx, req.Email)

	if existingUser != nil {
		logrus.Error("User already exists")
		return nil, errorEntity.ErrUserAlreadyExists
	}

	hashedPassword := util.HashPassword(req.Password, a.cfg.Server.Salt)

	newUser, err := entity.NewUser(
		req.Email,
		hashedPassword,
		req.Name,
		"",
	)

	if err != nil {
		logrus.Error("Failed to create new user")
		return nil, err
	}

	err = a.repo.Create(ctx, newUser)
	if err != nil {
		logrus.Error("Failed to create new user in database")
		return nil, err
	}

	logrus.Info("User created successfully")
	return newUser, nil
}

func (a *AuthUseCase) Login(req *dto.LoginRequest, ctx context.Context) (*dto.LoginResponse, error) {
	existingUser, err := a.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		logrus.Error("User not found")
		return nil, errorEntity.ErrUserNotFound
	}

	if existingUser.EmailVerified == nil {
		logrus.Error("Email not verified")
		return nil, errorEntity.ErrEmailNotVerified
	}

	if !util.ComparePassword(existingUser.Password, req.Password, a.cfg.Server.Salt) {
		logrus.Error("Invalid password")
		return nil, errorEntity.ErrInvalidPassword
	}

	acc, refresh, refreshHash, err := a.jwt.GenerateToken(existingUser.UUID, existingUser.Email)
	if err != nil {
		logrus.Error("Failed to generate token")
		return nil, err
	}

	time, _ := tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
	jsonUser, _ := json.Marshal(existingUser)
	err = a.redis.Set(fmt.Sprintf("user:%s", existingUser.UUID), jsonUser, time)
	if err != nil {
		logrus.Error("Failed to set access token in Redis")
		return nil, err
	}

	time, _ = tm.ParseDuration(a.cfg.Jwt.RefreshTokenExpire)
	err = a.redis.Set(fmt.Sprintf("user_refresh:%s", existingUser.UUID), refreshHash, time)
	if err != nil {
		logrus.Error("Failed to set refresh token in Redis")
		return nil, err
	}

	logrus.Info("User logged in successfully")
	return &dto.LoginResponse{
		AccessToken:  acc,
		RefreshToken: refresh,
	}, nil
}

func (a *AuthUseCase) Logout(req *dto.LogoutRequest, user *entity.User, accessToken string, ctx context.Context) error {
	claims, err := a.jwt.ValidateToken(req.RefreshToken, a.cfg.Jwt.RefreshTokenSecret)
	if err != nil {
		logrus.Error("Invalid refresh token")
		return errorEntity.ErrInvalidToken
	}

	if claims.UUID != user.UUID {
		logrus.Error("Invalid user")
		return errorEntity.ErrInvalidUser
	}

	err = a.redis.Delete(fmt.Sprintf("user:%s", user.UUID))
	if err != nil {
		logrus.Error("Failed to delete access token from Redis")
		return err
	}

	err = a.redis.Delete(fmt.Sprintf("user_refresh:%s", user.UUID))
	if err != nil {
		logrus.Error("Failed to delete refresh token from Redis")
		return err
	}

	isBlacklisted, err := a.redis.Exists(fmt.Sprintf("blacklist:%s", req.RefreshToken))
	if err != nil {
		logrus.Error("Failed to check if token is blacklisted")
		return err
	}

	if isBlacklisted {
		logrus.Error("Token is already blacklisted")
		return errorEntity.ErrTokenAlreadyBlacklisted
	}

	errCh := make(chan error, 2)

	go func() {
		expireDuration, _ := tm.ParseDuration(a.cfg.Jwt.RefreshTokenExpire)
		errCh <- a.redis.Set(fmt.Sprintf("blacklist:%s", req.RefreshToken), "blacklisted", expireDuration)
	}()

	go func() {
		expireDuration, _ := tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
		errCh <- a.redis.Set(fmt.Sprintf("blacklist:%s", accessToken), "blacklisted", expireDuration)
	}()

	var combinedErr error
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			logrus.Error("Failed to set blacklist key:", err)
			combinedErr = err
		}
	}
	close(errCh)

	if combinedErr != nil {
		logrus.Error("Failed to set blacklist key")
		return combinedErr
	}

	logrus.Info("User logged out successfully")

	return nil
}

func (a *AuthUseCase) RefreshToken(req *dto.RefreshTokenRequest, user *entity.User) (*dto.RefreshTokenResponse, error) {
	isBlacklisted, err := a.redis.Exists(fmt.Sprintf("blacklist:%s", req.RefreshToken))
	if err != nil {
		logrus.Error("Failed to check if token is blacklisted")
		return nil, err
	}

	if isBlacklisted {
		logrus.Error("Token is blacklisted")
		return nil, errorEntity.ErrTokenAlreadyBlacklisted
	}

	claims, err := a.jwt.ValidateToken(req.RefreshToken, a.cfg.Jwt.RefreshTokenSecret)
	if err != nil {
		logrus.Error("Invalid refresh token")
		return nil, errorEntity.ErrInvalidToken
	}

	if claims.UUID != user.UUID {
		logrus.Error("Invalid user")
		return nil, errorEntity.ErrInvalidUser
	}

	hashRefresh, err := a.redis.Get(fmt.Sprintf("user_refresh:%s", user.UUID))
	if err != nil {
		logrus.Error("Failed to get refresh token from Redis")
		return nil, err
	}

	err = a.jwt.CompareTokenHash(req.RefreshToken, hashRefresh)
	if err != nil {
		logrus.Error("Invalid refresh token")
		return nil, errorEntity.ErrInvalidToken
	}

	time, _ := tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
	token, err := a.jwt.GenerateSingleToken(claims.UUID, claims.Email, time, a.cfg.Jwt.AccessTokenSecret)
	if err != nil {
		logrus.Error("Failed to generate new access token")
		return nil, err
	}

	return &dto.RefreshTokenResponse{AccessToken: token}, nil
}

func (a *AuthUseCase) SendOtp(body *dto.SendOtpRequest, ctx context.Context) error {
	existingUser, err := a.repo.FindByEmail(ctx, body.Email)
	if err != nil {
		logrus.Error("User not found")
		return errorEntity.ErrUserNotFound
	}

	limitKey := fmt.Sprintf("otp_limit:%s", existingUser.UUID)

	count, err := a.redis.Incr(limitKey)
	if err != nil {
		logrus.WithError(err).Error("Failed to increment OTP limit key")
		return fmt.Errorf("internal error")
	}

	if count == 1 {
		err = a.redis.Expire(limitKey, 5*tm.Minute)
		if err != nil {
			logrus.WithError(err).Error("Failed to set expiry on OTP limit key")
			return fmt.Errorf("internal error")
		}
	}

	if count > 5 {
		logrus.Warn("OTP request limit exceeded")
		return errorEntity.ErrLimitExceeded
	}

	otp := util.GenerateOTP()

	go func(email, otp string) {
		err := a.mail.SendMail(email, "OTP Verification", fmt.Sprintf("Your OTP is: %s, This will expired after 5 minutes", otp))
		if err != nil {
			logrus.WithError(err).Error("Failed to send OTP email")
		}
	}(existingUser.Email, otp)

	err = a.redis.Set(fmt.Sprintf("otp:%s", existingUser.UUID), otp, 5*tm.Minute)
	if err != nil {
		logrus.Error("Failed to set OTP in Redis")
		return err
	}

	logrus.Info("OTP sent successfully")
	return nil
}

func (a *AuthUseCase) VerifyEmail(req *dto.VerifyEmailRequest, ctx context.Context) error {
	existingUser, err := a.repo.FindByEmail(ctx, req.Email)
	if err != nil {
		logrus.Error("User not found")
		return errorEntity.ErrUserNotFound
	}

	otp, err := a.redis.Get(fmt.Sprintf("otp:%s", existingUser.UUID))
	if err != nil {
		logrus.Error("Failed to get OTP from Redis")
		return errorEntity.ErrOtpNotFound
	}

	if otp != req.Otp {
		logrus.Error("Invalid OTP")
		return errorEntity.ErrInvalidOtp
	}

	err = a.repo.UpdateEmailVerified(ctx, existingUser.UUID)
	if err != nil {
		logrus.Error("Failed to update email verified status")
		return err
	}

	errCh := make(chan error, 2)

	go func() {
		errCh <- a.redis.Delete(fmt.Sprintf("otp:%s", existingUser.UUID))
	}()

	go func() {
		errCh <- a.redis.Delete(fmt.Sprintf("otp_limit:%s", existingUser.UUID))
	}()

	var combinedErr error
	for i := 0; i < 2; i++ {
		if err := <-errCh; err != nil {
			logrus.Error("Failed to delete OTP related key from Redis:", err)
			combinedErr = err
		}
	}

	if combinedErr != nil {
		logrus.Error("Failed to delete OTP related key from Redis")
		return combinedErr
	}

	close(errCh)

	logrus.Info("Email verified successfully")
	return nil
}
