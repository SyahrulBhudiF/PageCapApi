package handler

import (
	"errors"
	"fmt"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
	"github.com/markbates/goth/gothic"
)

type AuthHandler struct {
	auth *usecase.AuthUseCase
}

func NewAuthHandler(auth *usecase.AuthUseCase) *AuthHandler {
	return &AuthHandler{
		auth: auth,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Register a new user with name, email, and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param register body dto.RegisterRequest true "Register Request"
// @Success 201 {object} response.Response{data=entity.User} "user registered successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router /auth/register [post]
func (h *AuthHandler) Register(c *gin.Context) {
	req, err := util.GetBody[dto.RegisterRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	newUser, err := h.auth.Register(&req, c.Request.Context())
	if err != nil {
		if errors.Is(err, errorEntity.ErrUserAlreadyExists) {
			response.Conflict(c, "conflict request", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.Created(c, "user registered successfully", newUser)
}

// Login godoc
// @Summary Login user
// @Description Login user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param login body dto.LoginRequest true "Login Request"
// @Success 201 {object} response.Response{data=dto.LoginResponse} "user logged in successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	req, err := util.GetBody[dto.LoginRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	token, err := h.auth.Login(&req, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrInvalidPassword, errorEntity.ErrUserNotFound, errorEntity.ErrEmailNotVerified) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "user logged in successfully", token)
}

// Logout godoc
// @Summary Logout user
// @Description Logout user and invalidate the refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param logout body dto.LogoutRequest true "Logout Request"
// @Success 200 {object} response.Response "user logged out successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Security BearerAuth
// @Router /auth/logout [delete]
func (h *AuthHandler) Logout(c *gin.Context) {
	body, err := util.GetBody[dto.LogoutRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	user, err := util.GetBody[entity.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	accessTokenRaw, exists := c.Get("accessToken")
	if !exists {
		response.BadRequest(c, "invalid request", fmt.Errorf("accessToken not found"))
		return
	}

	accessToken, ok := accessTokenRaw.(string)
	if !ok {
		response.BadRequest(c, "invalid request", fmt.Errorf("invalid accessToken type"))
		return
	}

	err = h.auth.Logout(&body, &user, accessToken)
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrInvalidToken, errorEntity.ErrInvalidUser, errorEntity.ErrTokenAlreadyBlacklisted) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "user logged out successfully", nil)
}

// RefreshToken godoc
// @Summary Refresh access token
// @Description Refresh access token using refresh token
// @Tags Auth
// @Accept json
// @Produce json
// @Param refresh body dto.RefreshTokenRequest true "Refresh Token Request"
// @Success 200 {object} response.Response{data=dto.RefreshTokenResponse} "access token refreshed successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Security BearerAuth
// @Router /auth/refresh [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	body, err := util.GetBody[dto.RefreshTokenRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	user, err := util.GetBody[entity.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	token, err := h.auth.RefreshToken(&body, &user)
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrInvalidToken, errorEntity.ErrTokenAlreadyBlacklisted, errorEntity.ErrInvalidUser) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "access token refreshed successfully", token)
}

// SendOtp godoc
// @Summary Send OTP
// @Description Send OTP to user's email
// @Tags Auth
// @Accept json
// @Produce json
// @Param sendOtp body dto.SendOtpRequest true "Send OTP Request"
// @Success 200 {object} response.Response "OTP sent successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router /auth/send-otp [post]
func (h *AuthHandler) SendOtp(c *gin.Context) {
	body, err := util.GetBody[dto.SendOtpRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	err = h.auth.SendOtp(&body, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrUserNotFound) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "OTP sent successfully", nil)
}

// VerifyEmail godoc
// @Summary Verify email
// @Description Verify email using OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param verifyEmail body dto.VerifyEmailRequest true "Verify Email Request"
// @Success 200 {object} response.Response "Email verified successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router /auth/verify-email [post]
func (h *AuthHandler) VerifyEmail(c *gin.Context) {
	body, err := util.GetBody[dto.VerifyEmailRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	err = h.auth.VerifyEmail(&body, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrInvalidOtp, errorEntity.ErrUserNotFound) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Email verified successfully", nil)
}

// ForgotPassword godoc
// @Summary Forgot password
// @Description Reset password using OTP
// @Tags Auth
// @Accept json
// @Produce json
// @Param forgotPassword body dto.ForgotPasswordRequest true "Forgot Password Request"
// @Success 200 {object} response.Response "Password reset successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router /auth/forgot-password [post]
func (h *AuthHandler) ForgotPassword(c *gin.Context) {
	body, err := util.GetBody[dto.ForgotPasswordRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	err = h.auth.ForgotPassword(&body, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrInvalidOtp, errorEntity.ErrUserNotFound) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Password reset successfully", nil)
}

// GoogleLogin godoc
// @Summary Google login
// @Description Login using Google OAuth2
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response "Google login successful"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router /auth/google [get]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	c.Request = c.Request.WithContext(c)

	gothic.BeginAuthHandler(c.Writer, c.Request)
}

// GoogleCallback godoc
// @Summary Google callback
// @Description Callback URL for Google OAuth2
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.LoginResponse} "Google login successful"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Router /auth/google/callback [get]
func (h *AuthHandler) GoogleCallback(c *gin.Context) {
	c.Request = c.Request.WithContext(c)

	user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
	if err != nil {
		response.Unauthorized(c, "unauthorized", err)
		return
	}

	token, err := h.auth.GoogleLogin(&user, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrUserNotFound, errorEntity.ErrEmailNotVerified) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Google login successful", token)
}

// SetPassword godoc
// @Summary Set password
// @Description Set password for the user after Google login
// @Tags Auth
// @Accept json
// @Produce json
// @Param setPassword body dto.SetPasswordRequest true "Set Password Request"
// @Success 200 {object} response.Response "Password set successfully"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Security BearerAuth
// @Router /auth/set-password [post]
func (h *AuthHandler) SetPassword(c *gin.Context) {
	body, err := util.GetBody[dto.SetPasswordRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	user, err := util.GetBody[entity.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	err = h.auth.SetPassword(&body, &user, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrUserNotFound, errorEntity.ErrUserAlreadyHasPassword) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Password set successfully", nil)
}

// GenerateApiKey godoc
// @Summary Generate API Key
// @Description Generate an API key for user
// @Tags Auth
// @Accept json
// @Produce json
// @Success 200 {object} response.Response{data=dto.ApiKeyResponse} "Successfully generate API key"
// @Failure 400 {object} response.ErrorResponse "invalid request"
// @Failure 401 {object} response.ErrorResponse "unauthorized"
// @Failure 500 {object} response.ErrorResponse "internal server error"
// @Security BearerAuth
// @Router /auth/api-key [get]
func (h *AuthHandler) GenerateApiKey(c *gin.Context) {
	user, err := util.GetBody[entity.User](c, "user")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	apiKey, err := h.auth.GenerateApiKey(&user)
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrUserNotFound) {
			response.Unauthorized(c, "unauthorized", err)
		} else {
			response.InternalServerError(c, err)
		}
		return
	}

	response.OK(c, "Successfully generate API key", apiKey)
}
