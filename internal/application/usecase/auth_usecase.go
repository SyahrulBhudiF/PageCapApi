package usecase

import (
	"context"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/mail"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/sirupsen/logrus"
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

func (a *AuthUseCase) Register(user *entity.User, ctx context.Context) (*entity.User, error) {
	existingUser, err := a.repo.FindByEmail(ctx, user.Email)
	if err != nil {
		return nil, err
	}

	if existingUser != nil {
		logrus.Error("User already exists")
		return nil, errorEntity.ErrUserAlreadyExists
	}

	hashedPassword := util.HashPassword(user.Password, a.cfg.Server.Salt)

	newUser, err := entity.NewUser(
		user.Email,
		user.Name,
		hashedPassword,
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
