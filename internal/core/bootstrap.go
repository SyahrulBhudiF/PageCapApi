package core

import (
	docs "github.com/SyahrulBhudiF/Doc-Management.git/docs"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/core/module"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/cloudinary"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/database"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/jwt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/mail"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/oauth2/google"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/persistence"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/redis"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/infrastructure/rod"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/midleware"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/route"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/config"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"net/http"
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

	// Initialize Oauth2 providers
	google.NewGoogle(cfg)
	gothic.GetProviderName = func(r *http.Request) (string, error) {
		return "google", nil
	}

	// Cloudinary
	cloudinaryService := cloudinary.NewCloudinary(cfg)

	// Init Browser
	browser, err := rod.InitBrowser()
	if err != nil {
		return nil, err
	}

	// Repositories
	userRepo := persistence.NewUserRepository(db)
	pageCaptureRepo := persistence.NewPageCaptureRepository(db)

	// Initialize middleware
	authMiddleware := midleware.NewAuthMiddleware(userRepo, redisRepo, jwtService, cfg)

	// Initialize Modules
	authHandler := module.InitAuthModule(cfg, userRepo, jwtService, mailService, redisRepo)
	userHandler := module.InitUserModule(cfg, userRepo, redisRepo, cloudinaryService)
	pageCaptureHandler := module.InitPageCaptureModule(cfg, pageCaptureRepo, redisRepo, cloudinaryService, browser)

	// Router
	docs.SwaggerInfo.BasePath = "/api/v1"
	r := route.NewRoute(authHandler, authMiddleware, userHandler, pageCaptureHandler)
	router := r.RegisterRoutes()
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	return &App{Router: router}, nil
}
