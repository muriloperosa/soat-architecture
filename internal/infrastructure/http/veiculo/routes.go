package veiculo

import (
	"github.com/gin-gonic/gin"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

func RegisterVeiculoRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	h := NewHandler(c.CadastrarVeiculoUC, c.AtualizarVeiculoUC, c.AtivarVeiculoUC, c.InativarVeiculoUC, c.ConsultarVeiculoPorIDUC, c.ConsultarVeiculoPorPlacaUC, c.ListarVeiculosUC, httpquery.NewParser())

	autenticado := rg.Group("/veiculos", middleware.AuthenticationMiddleware(c.JWTAuth, c.RefreshTokensRepo, c.UsuarioStatusRepo, c.ClienteStatusRepo))

	gestao := autenticado.Group("", middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelAdmin))
	gestao.POST("", h.Cadastrar)
	gestao.PUT("/:id", h.Atualizar)
	gestao.PATCH("/:id/ativar", h.Ativar)
	gestao.PATCH("/:id/inativar", h.Inativar)

	consulta := autenticado.Group("", middleware.AuthorizationMiddleware(domainauth.TipoInterno))
	consulta.GET("/placa/:placa", h.ConsultarPorPlaca)
	consulta.GET("/:id", h.ConsultarPorID)
	consulta.GET("", h.Listar)
}
