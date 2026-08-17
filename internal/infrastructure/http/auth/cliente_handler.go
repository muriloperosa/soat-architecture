package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

// AuthClienteHandler expõe login/refresh/logout pro cliente.
type AuthClienteHandler struct {
	login   *appauth.LoginUseCase
	refresh *appauth.RefreshUseCase
	logout  *appauth.LogoutUseCase
}

func NewAuthClienteHandler(login *appauth.LoginUseCase, refresh *appauth.RefreshUseCase, logout *appauth.LogoutUseCase) *AuthClienteHandler {
	return &AuthClienteHandler{login: login, refresh: refresh, logout: logout}
}

// @Summary Login de cliente
// @Description Autentica por email+senha e emite o par access+refresh token
// @Tags Auth Cliente
// @Accept json
// @Produce json
// @Param request body LoginRequest true "Credenciais"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Router /v1/auth/cliente/login [post]
func (h *AuthClienteHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.login.Executar(c.Request.Context(), toLoginInput(req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	resp := toTokenResponse(out.AccessToken, out.RefreshToken, out.AccessTokenExpiresIn, out.RefreshTokenExpiresIn, out.RequerAlterarSenha)
	c.JSON(http.StatusOK, resp)
}

// @Summary Refresh de token de cliente
// @Description Troca um refresh token válido por um novo par access+refresh (rotação)
// @Tags Auth Cliente
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 200 {object} TokenResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Router /v1/auth/cliente/refresh [post]
func (h *AuthClienteHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.refresh.Executar(c.Request.Context(), toRefreshInput(req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTokenResponse(out.AccessToken, out.RefreshToken, out.AccessTokenExpiresIn, out.RefreshTokenExpiresIn, false))
}

// @Summary Logout de cliente
// @Description Revoga o refresh token informado (idempotente)
// @Tags Auth Cliente
// @Accept json
// @Produce json
// @Param request body RefreshRequest true "Refresh token"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Router /v1/auth/cliente/logout [post]
func (h *AuthClienteHandler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	if err := h.logout.Executar(c.Request.Context(), toLogoutInput(req)); err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
