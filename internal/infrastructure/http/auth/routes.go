package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterAuthRoutes registra login/refresh/logout do usuário interno.
// Cliente (/v1/auth/cliente/*) permanece fora de escopo — depende do
// domínio cliente, que ainda não existe.
func RegisterAuthRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	interno := NewAuthInternoHandler(c.LoginInternoUC, c.RefreshUC, c.LogoutUC)
	rg.POST("/auth/login", interno.Login)
	rg.POST("/auth/refresh", interno.Refresh)
	rg.POST("/auth/logout", interno.Logout)
}
