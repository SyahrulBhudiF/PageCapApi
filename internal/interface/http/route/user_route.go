package route

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/midleware"
	"github.com/gin-gonic/gin"
)

func RegisterUserRoutes(rg *gin.RouterGroup, userHandler *handler.UserHandler, mm *midleware.AuthMiddleware) {
	user := rg.Group("/user")
	{
		user.GET("/profile", mm.EnsureAuthenticated(), userHandler.GetProfile)
		user.PATCH("/change-password", mm.EnsureAuthenticated(), midleware.EnsureJsonValidRequest[dto.ChangePasswordRequest](), userHandler.ChangePassword)
	}
}
