package http

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
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

func (h *AuthClienteHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"type": "validation", "message": "corpo inválido"})
		return
	}

	out, err := h.login.Executar(c.Request.Context(), toLoginInput(req))
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTokenResponse(out.AccessToken, out.RefreshToken))
}

func (h *AuthClienteHandler) Refresh(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"type": "validation", "message": "corpo inválido"})
		return
	}

	out, err := h.refresh.Executar(c.Request.Context(), toRefreshInput(req))
	if err != nil {
		RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toTokenResponse(out.AccessToken, out.RefreshToken))
}

func (h *AuthClienteHandler) Logout(c *gin.Context) {
	var req RefreshRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"type": "validation", "message": "corpo inválido"})
		return
	}

	if err := h.logout.Executar(c.Request.Context(), toLogoutInput(req)); err != nil {
		RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
