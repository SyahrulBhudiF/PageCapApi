package route

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interfaces/http/handler"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, authHandler *handler.AuthHandler) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", authHandler.Register)
	}
}
