package module

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	jwtContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/jwt"
	redisContract "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/contract/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/mail"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/persistence"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interfaces/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"gorm.io/gorm"
)

func InitAuthModule(cfg *config.Config, db *gorm.DB, jwtService jwtContract.Service, mailService *mail.Service, redisRepo redisContract.Service) *handler.AuthHandler {
	// Persistence
	userRepo := persistence.NewUserRepository(db)

	// HTTP
	authUC := usecase.NewAuthUseCase(userRepo, jwtService, redisRepo, mailService, cfg)
	authHandler := handler.NewAuthHandler(authUC)

	return authHandler
}
