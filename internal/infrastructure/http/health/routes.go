package health

import (
	"github.com/gin-gonic/gin"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterHealthRoutes registra as rotas de health check.
func RegisterHealthRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	rg.GET("/health", NewHealthHandler(c.DB))
}
