package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/muriloperosa/soat-architecture/docs/swagger"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/health"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// NewRouter monta o *gin.Engine e delega o registro de rotas por domínio.
func NewRouter(c *wiring.Container) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/v1")
	health.RegisterHealthRoutes(v1, c)

	return router
}
