package orcamento

import (
	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterOrcamentoRoutes registra as rotas de orçamento aninhadas em uma
// Ordem de Serviço, restritas a mecânico ou administrador.
func RegisterOrcamentoRoutes(rg *gin.RouterGroup, container *wiring.Container) {
	handler := NewHandler(
		container.GerarOrcamentoUC,
		container.AdicionarServicoOrcamentoUC,
		container.AdicionarPecaOrcamentoUC,
		container.RemoverServicoOrcamentoUC,
		container.RemoverPecaOrcamentoUC,
	)

	orcamentos := rg.Group(
		"/ordens-servico/:id/orcamento",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
		middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelMecanico, shared.PapelAdmin),
	)

	orcamentos.POST("", handler.Gerar)
	orcamentos.POST("/itens-servico", handler.AdicionarServico)
	orcamentos.POST("/itens-peca", handler.AdicionarPeca)
	orcamentos.DELETE("/itens-servico/:itemId", handler.RemoverServico)
	orcamentos.DELETE("/itens-peca/:itemId", handler.RemoverPeca)
}
