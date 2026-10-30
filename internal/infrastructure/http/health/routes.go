package health

import (
	"github.com/gin-gonic/gin"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterHealthRoutes registra as rotas de health check e as probes usadas
// pelo Kubernetes (live e ready).
func RegisterHealthRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	rg.GET("/health", NewHealthHandler(c.DB))
	rg.GET("/health/live", NewLivenessHandler())
	rg.GET("/health/ready", NewReadinessHandler(c.DB))
}
