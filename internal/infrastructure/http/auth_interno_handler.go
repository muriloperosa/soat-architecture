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

func (h *AuthInternoHandler) Login(c *gin.Context) {
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

func (h *AuthInternoHandler) Refresh(c *gin.Context) {
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

func (h *AuthInternoHandler) Logout(c *gin.Context) {
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
