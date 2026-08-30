package http

import (
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "github.com/muriloperosa/soat-architecture/docs/swagger"
	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
	httpcliente "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/cliente"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/health"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
	httppeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/peca"
	httprelatorio "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/relatorio"
	httpservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/servico"
	httpusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/usuario"
	httpveiculo "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

func NewRouter(c *wiring.Container) *gin.Engine {
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(middleware.RecoveryMiddleware())
	router.Use(middleware.SecurityHeadersMiddleware())

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	v1 := router.Group("/v1")
	health.RegisterHealthRoutes(v1, c)
	httpauth.RegisterAuthRoutes(v1, c)
	httpusuario.RegisterUsuarioRoutes(v1, c)
	httpcliente.RegisterClienteRoutes(v1, c)
	httppeca.RegisterPecaRoutes(v1, c)
	httpveiculo.RegisterVeiculoRoutes(v1, c)
	httpservico.RegisterServicoRoutes(v1, c)
	httpordemservico.RegisterOrdemServicoRoutes(v1, c)
	httporcamento.RegisterOrcamentoRoutes(v1, c)
	httprelatorio.RegisterRelatorioRoutes(v1, c)

	return router
}
