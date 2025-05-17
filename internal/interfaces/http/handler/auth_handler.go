package handler

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/entity"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/interfaces/http/dto"
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

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	newUser, err := h.auth.Register(&entity.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}, c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Created(c, "user registered successfully", newUser)
}
