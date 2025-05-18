package handler

import (
	"errors"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	errorEntity "github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/error"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/shared/util"
	"github.com/SyahrulBhudiF/Doc-Management.git/pkg/response"
	"github.com/gin-gonic/gin"
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
// @Failure 400 {object} response.Response "invalid request"
// @Failure 500 {object} response.Response "internal server error"
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
// @Failure 400 {object} response.Response "invalid request"
// @Failure 401 {object} response.Response "unauthorized"
// @Failure 500 {object} response.Response "internal server error"
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
	req, err := util.GetBody[dto.LoginRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	token, err := h.auth.Login(&req, c.Request.Context())
	if err != nil {
		if util.ErrorInList(err, errorEntity.ErrInvalidPassword, errorEntity.ErrUserNotFound) {
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
// @Failure 400 {object} response.Response "invalid request"
// @Failure 401 {object} response.Response "unauthorized"
// @Failure 500 {object} response.Response "internal server error"
// @Security BearerAuth
// @Router /auth/logout [delete]
func (h *AuthHandler) Logout(c *gin.Context) {
	body, err := util.GetBody[dto.LogoutRequest](c, "body")
	if err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	user, err := util.GetBody[entity.User](c, "user")
	accessToken, _ := c.Get("accessToken")

	err = h.auth.Logout(&body, &user, accessToken.(string), c.Request.Context())
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
