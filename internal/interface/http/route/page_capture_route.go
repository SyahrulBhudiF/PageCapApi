package route

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/midleware"
	"github.com/gin-gonic/gin"
)

func RegisterPageCaptureRoutes(rg *gin.RouterGroup, authHandler *handler.PageCaptureHandler, mm *midleware.AuthMiddleware) {
	r := rg.Group("/page-capture")
	{
		r.POST("/:key", midleware.EnsureJsonValidRequest[dto.PageCaptureRequest](), authHandler.PageCapture)
	}
}
