package module

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	redisContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/cloudinary/cloudinary-go/v2"
)

func InitUserModule(cfg *config.Config, repo repository.UserRepository, redis redisContract.Service, cloud *cloudinary.Cloudinary) *handler.UserHandler {
	userUC := usecase.NewUserUseCase(repo, redis, cfg, cloud)
	userHandler := handler.NewUserHandler(userUC)

	return userHandler
}
