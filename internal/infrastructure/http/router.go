package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	_ "github.com/muriloperosa/soat-architecture/docs/swagger"
)

// NewRouter monta o *gin.Engine com todas as rotas da aplicação.
func NewRouter(db *gorm.DB) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/v1")
	v1.GET("/health", NewHealthHandler(db))

	return router
}
