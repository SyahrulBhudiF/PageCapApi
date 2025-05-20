package route

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/handler"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interface/http/midleware"
	"github.com/gin-gonic/gin"
)

func RegisterAuthRoutes(rg *gin.RouterGroup, authHandler *handler.AuthHandler, mm *midleware.AuthMiddleware) {
	auth := rg.Group("/auth")
	{
		auth.POST("/register", midleware.EnsureJsonValidRequest[dto.RegisterRequest](), authHandler.Register)
		auth.POST("/login", midleware.EnsureJsonValidRequest[dto.LoginRequest](), authHandler.Login)
		auth.DELETE("/logout", mm.EnsureAuthenticated(), midleware.EnsureJsonValidRequest[dto.LogoutRequest](), authHandler.Logout)
		auth.POST("/refresh", mm.EnsureAuthenticated(), midleware.EnsureJsonValidRequest[dto.RefreshTokenRequest](), authHandler.RefreshToken)
		auth.POST("/send-otp", midleware.EnsureJsonValidRequest[dto.SendOtpRequest](), authHandler.SendOtp)
		auth.POST("/verify-email", midleware.EnsureJsonValidRequest[dto.VerifyEmailRequest](), authHandler.VerifyEmail)
		auth.POST("/forgot-password", midleware.EnsureJsonValidRequest[dto.ForgotPasswordRequest](), authHandler.ForgotPassword)
		auth.GET("/google", authHandler.GoogleLogin)
		auth.GET("/google/callback", authHandler.GoogleCallback)
		auth.POST("/set-password", mm.EnsureAuthenticated(), midleware.EnsureJsonValidRequest[dto.SetPasswordRequest](), authHandler.SetPassword)
		auth.GET("/api-key", mm.EnsureAuthenticated(), authHandler.GenerateApiKey)
	}
}
