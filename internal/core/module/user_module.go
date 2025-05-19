package module

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	redisContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
)

func InitUserModule(cfg *config.Config, repo repository.UserRepository, redis redisContract.Service) *handler.UserHandler {
	userUC := usecase.NewUserUseCase(repo, redis, cfg)
	userHandler := handler.NewUserHandler(userUC)

	return userHandler
}
