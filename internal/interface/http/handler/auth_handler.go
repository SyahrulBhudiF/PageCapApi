package handler

import (
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/application/usecase"
	"github.com/SyahrulBhudiF/Doc-Management.git/internal/domain/dto"
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
	var req dto.RegisterRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	if err := req.Validate(); err != nil {
		response.BadRequest(c, "invalid request", err)
		return
	}

	newUser, err := h.auth.Register(&req, c.Request.Context())
	if err != nil {
		response.InternalServerError(c, err)
		return
	}

	response.Created(c, "user registered successfully", newUser)
}
