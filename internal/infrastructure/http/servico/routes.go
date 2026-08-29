package servico

import (
	"github.com/gin-gonic/gin"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterServicoRoutes registra o CRUD do catálogo de serviços, restrito a
// usuário interno autenticado (qualquer papel).
func RegisterServicoRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	h := NewHandler(
		c.CriarServicoUC,
		c.AtualizarServicoUC,
		c.ListarServicosUC,
		c.BuscarServicoUC,
		c.AtivarServicoUC,
		c.InativarServicoUC,
		httpquery.NewParser(),
	)

	g := rg.Group("/servicos",
		middleware.AuthenticationMiddleware(c.JWTAuth, c.RefreshTokensRepo, c.UsuarioStatusRepo, c.ClienteStatusRepo),
		middleware.AuthorizationMiddleware(domainauth.TipoInterno),
	)
	g.POST("", h.Criar)
	g.GET("", h.Listar)
	g.GET("/:id", h.Buscar)
	g.PUT("/:id", h.Atualizar)
	g.PATCH("/:id/ativar", h.Ativar)
	g.PATCH("/:id/inativar", h.Inativar)
}
