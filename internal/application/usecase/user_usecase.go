package usecase

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
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
