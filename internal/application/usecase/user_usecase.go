package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/sirupsen/logrus"
	tm "time"
)

type UserUseCase struct {
	repo  repository.UserRepository
	redis redis.Service
	cfg   *config.Config
}

func NewUserUseCase(repo repository.UserRepository, redis redis.Service, cfg *config.Config) *UserUseCase {
	return &UserUseCase{
		repo:  repo,
		redis: redis,
		cfg:   cfg,
	}
}

func (c *UserUseCase) ChangePassword(d *dto.ChangePasswordRequest, e *entity.User, context context.Context) error {
	if e.Password == "" {
		logrus.WithError(errorEntity.ErrPasswordNotSet).Error("password is not set")
		return errorEntity.ErrPasswordNotSet
	}

	if response := util.ComparePassword(e.Password, d.OldPassword, c.cfg.Server.Salt); !response {
		logrus.Info("password not match")
		return errorEntity.ErrInvalidPassword
	}

	hashedPassword := util.HashPassword(d.NewPassword, c.cfg.Server.Salt)
	e.Password = hashedPassword
	if err := c.repo.Update(context, e); err != nil {
		logrus.WithError(err).Error("failed to update password")
		return err
	}

	if err := c.redis.Delete(fmt.Sprintf("user:%s", e.UUID)); err != nil {
		logrus.WithError(err).Error("failed to delete user cache")
		return err
	}

	accessExpire, _ := tm.ParseDuration(c.cfg.Jwt.AccessTokenExpire)
	jsonUser, _ := json.Marshal(e)
	if err := c.redis.Set(fmt.Sprintf("user:%s", e.UUID), jsonUser, accessExpire); err != nil {
		logrus.WithError(err).Error("failed to set user cache")
		return err
	}

	logrus.Info("password updated successfully")
	return nil
}
