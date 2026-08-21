package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterAuthRoutes registra login/refresh/logout de usuários internos e clientes.
func RegisterAuthRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	interno := NewAuthInternoHandler(c.LoginInternoUC, c.RefreshUC, c.LogoutUC)
	rg.POST("/auth/login", interno.Login)
	rg.POST("/auth/refresh", interno.Refresh)
	rg.POST("/auth/logout", interno.Logout)

	cliente := NewAuthClienteHandler(c.LoginClienteUC, c.RefreshUC, c.LogoutUC)
	rg.POST("/auth/cliente/login", cliente.Login)
	rg.POST("/auth/cliente/refresh", cliente.Refresh)
	rg.POST("/auth/cliente/logout", cliente.Logout)
}
