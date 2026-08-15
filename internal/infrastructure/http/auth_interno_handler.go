package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
)

// AuthInternoHandler expõe login/refresh/logout pro usuário interno.
type AuthInternoHandler struct {
	login   *appauth.LoginUseCase
	refresh *appauth.RefreshUseCase
	logout  *appauth.LogoutUseCase
}

func NewAuthInternoHandler(login *appauth.LoginUseCase, refresh *appauth.RefreshUseCase, logout *appauth.LogoutUseCase) *AuthInternoHandler {
	return &AuthInternoHandler{login: login, refresh: refresh, logout: logout}
}

// @Summary Login de usuário interno
// @Description Autentica por email+senha e emite o par access+refresh token
// @Tags auth-interno
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciais"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/auth/login [post]
func (h *AuthInternoHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, "corpo inválido")
		return
	}

	out, err := h.login.Executar(c.Request.Context(), toLoginInput(req))
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTokenResponse(out.AccessToken, out.RefreshToken))
}

// @Summary Refresh de token de usuário interno
// @Description Troca um refresh token válido por um novo par access+refresh (rotação)
// @Tags auth-interno
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} ErrorResponse
// @Failure 401 {object} ErrorResponse
// @Router /v1/auth/refresh [post]
func (h *AuthInternoHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, "corpo inválido")
		return
	}

	out, err := h.refresh.Executar(c.Request.Context(), toRefreshInput(req))
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTokenResponse(out.AccessToken, out.RefreshToken))
}

// @Summary Logout de usuário interno
// @Description Revoga o refresh token informado (idempotente)
// @Tags auth-interno
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} ErrorResponse
// @Router /v1/auth/logout [post]
func (h *AuthInternoHandler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		RespondValidationError(c, "corpo inválido")
		return
	}

	if err := h.logout.Executar(c.Request.Context(), toLogoutInput(req)); err != nil {
		RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
