package orcamento

import (
	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterOrcamentoRoutes registra as rotas de orçamento aninhadas em uma
// Ordem de Serviço. Edição/envio são internos; aprovação/rejeição são do
// cliente proprietário da OS, com a propriedade revalidada na application.
func RegisterOrcamentoRoutes(rg *gin.RouterGroup, container *wiring.Container) {
	handler := NewHandler(
		container.GerarOrcamentoUC,
		container.AdicionarServicoOrcamentoUC,
		container.AdicionarPecaOrcamentoUC,
		container.RemoverServicoOrcamentoUC,
		container.RemoverPecaOrcamentoUC,
		container.EnviarOrcamentoParaAprovacaoUC,
	)
	decisaoHandler := NewDecisaoHandler(container.AprovarOrcamentoUC, container.RejeitarOrcamentoUC)

	orcamentos := rg.Group(
		"/ordens-servico/:id/orcamento",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
	)

	internos := orcamentos.Group(
		"",
		middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelMecanico, shared.PapelAdmin),
	)
	internos.POST("", handler.Gerar)
	internos.POST("/itens-servico", handler.AdicionarServico)
	internos.POST("/itens-peca", handler.AdicionarPeca)
	internos.DELETE("/itens-servico/:itemId", handler.RemoverServico)
	internos.DELETE("/itens-peca/:itemId", handler.RemoverPeca)
	internos.PATCH("/enviar-aprovacao", handler.EnviarParaAprovacao)

	clientes := orcamentos.Group(
		"",
		middleware.AuthorizationMiddleware(domainauth.TipoCliente),
	)
	clientes.PATCH("/aprovar", decisaoHandler.Aprovar)
	clientes.PATCH("/rejeitar", decisaoHandler.Rejeitar)
}
