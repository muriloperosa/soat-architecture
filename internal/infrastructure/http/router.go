package http

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// NewRouter monta o *gin.Engine com todas as rotas da aplicação.
func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.Default()

	router.GET("/health", NewHealthHandler(db))

	return router
}
