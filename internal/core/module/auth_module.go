package module

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	jwtContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	redisContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/repository"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/mail"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
)

func InitAuthModule(cfg *config.Config, repo repository.UserRepository, jwtService jwtContract.Service, mailService *mail.Service, redisRepo redisContract.Service) *handler.AuthHandler {
	// HTTP
	authUC := usecase.NewAuthUseCase(repo, jwtService, redisRepo, mailService, cfg)
	authHandler := handler.NewAuthHandler(authUC)

	return authHandler
}
