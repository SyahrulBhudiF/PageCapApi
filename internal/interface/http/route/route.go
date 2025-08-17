package route

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/midleware"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Route struct {
	AuthHandler        *handler.AuthHandler
	AuthMiddleware     *midleware.AuthMiddleware
	UserHandler        *handler.UserHandler
	PageCaptureHandler *handler.PageCaptureHandler
}

func NewRoute(authHandler *handler.AuthHandler, middleware *midleware.AuthMiddleware, UserHandler *handler.UserHandler, PageHandler *handler.PageCaptureHandler) *Route {
	return &Route{
		AuthHandler:        authHandler,
		AuthMiddleware:     middleware,
		UserHandler:        UserHandler,
		PageCaptureHandler: PageHandler,
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
		v1.HEAD("/health", func(c *gin.Context) {
			c.String(http.StatusOK, "ok")
		})
		// Auth
		RegisterAuthRoutes(v1, r.AuthHandler, r.AuthMiddleware)
		// User
		RegisterUserRoutes(v1, r.UserHandler, r.AuthMiddleware)
		// Page Capture
		RegisterPageCaptureRoutes(v1, r.PageCaptureHandler, r.AuthMiddleware)
	}

	return router
}
