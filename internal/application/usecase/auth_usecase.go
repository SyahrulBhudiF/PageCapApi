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

	expireDuration, _ := tm.ParseDuration(a.cfg.Jwt.RefreshTokenExpire)
	err = a.redis.Set(fmt.Sprintf("blacklist:%s", req.RefreshToken), "blacklisted", expireDuration)

	expireDuration, _ = tm.ParseDuration(a.cfg.Jwt.AccessTokenExpire)
	err = a.redis.Set(fmt.Sprintf("blacklist:%s", accessToken), "blacklisted", expireDuration)

	logrus.Info("User logged out successfully")

	return nil
}
