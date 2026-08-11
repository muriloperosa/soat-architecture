package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewHealthHandler retorna um handler Gin que faz ping no banco via GORM.
func NewHealthHandler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		sqlDB, err := db.DB()
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "erro", "detail": err.Error()})
			return
		}

		if err := sqlDB.PingContext(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "erro", "detail": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	}
}
