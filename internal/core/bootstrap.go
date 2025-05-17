package core

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/core/module"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/database"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/jwt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/mail"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interfaces/http/route"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/gin-gonic/gin"
)

type App struct {
	Router *gin.Engine
}

func Bootstrap() (*App, error) {
	cfg, err := config.Load()
	if err != nil {
		return nil, err
	}

	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		return nil, err
	}

	rd, err := redis.NewRedis(cfg)
	if err != nil {
		return nil, err
	}

	// Infrastructure services (reusable)
	jwtService := jwt.NewJwtService(cfg)
	mailService := mail.NewMailService(cfg)
	redisRepo := redis.NewRedisService(rd, "client")

	// Initialize Modules
	authHandler := module.InitAuthModule(cfg, db, jwtService, mailService, redisRepo)

	// Router
	r := route.NewRoute(authHandler)
	router := r.RegisterRoutes()

	return &App{Router: router}, nil
}
