package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewHealthHandler retorna um handler Gin que faz ping no banco via GORM.
//
// @Summary Verifica saúde da API
// @Description Retorna 200 se a API e a conexão com o MySQL estão respondendo
// @Tags health
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 503 {object} map[string]string
// @Router /v1/health [get]
func NewHealthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "detail": err.Error()})
			return
		}

		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "error", "detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
