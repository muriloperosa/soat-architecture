package peca

import (
	"github.com/gin-gonic/gin"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterPecaRoutes registra as rotas de gestão de peças de estoque.
func RegisterPecaRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	h := NewHandler(c.CadastrarPecaUC, c.AtualizarPecaUC, c.AtivarPecaUC, c.InativarPecaUC, c.ConsultarPecaPorIDUC, c.ReporEstoquePecaUC)

	autenticado := rg.Group("/pecas", middleware.AuthenticationMiddleware(c.JWTAuth, c.RefreshTokensRepo, c.UsuarioStatusRepo))

	gestao := autenticado.Group("", middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelAdmin))
	gestao.POST("", h.Cadastrar)
	gestao.PUT("/:id", h.Atualizar)
	gestao.PATCH("/:id/ativar", h.Ativar)
	gestao.PATCH("/:id/inativar", h.Inativar)
	gestao.PATCH("/:id/repor-estoque", h.ReporEstoque)

	consulta := autenticado.Group("", middleware.AuthorizationMiddleware(domainauth.TipoInterno))
	consulta.GET("/:id", h.ConsultarPorID)
}
