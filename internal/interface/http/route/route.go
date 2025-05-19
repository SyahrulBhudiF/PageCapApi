package route

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/midleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Route struct {
	AuthHandler    *handler.AuthHandler
	AuthMiddleware *midleware.AuthMiddleware
	UserHandler    *handler.UserHandler
}

func NewRoute(authHandler *handler.AuthHandler, middleware *midleware.AuthMiddleware, UserHandler *handler.UserHandler) *Route {
	return &Route{
		AuthHandler:    authHandler,
		AuthMiddleware: middleware,
		UserHandler:    UserHandler,
	}
}

func (r *Route) RegisterRoutes() *gin.Engine {
	gin.SetMode(gin.DebugMode)

	router := gin.New()

	router.Use(cors.New(CorsConfig()))
	router.Use(CustomRecovery())
	router.Use(Logger())
	router.Use(gin.Recovery())

	v1 := router.Group("/api/v1")
	{
		// Auth
		RegisterAuthRoutes(v1, r.AuthHandler, r.AuthMiddleware)
		// User
		RegisterUserRoutes(v1, r.UserHandler, r.AuthMiddleware)
	}

	return router
}
