package route

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interfaces/http/handler"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Route struct {
	AuthHandler *handler.AuthHandler
}

func NewRoute(authHandler *handler.AuthHandler) *Route {
	return &Route{
		AuthHandler: authHandler,
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
		RegisterAuthRoutes(v1, r.AuthHandler)
	}

	return router
}
