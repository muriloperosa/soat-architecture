package usuario

import (
	"github.com/gin-gonic/gin"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterUsuarioRoutes registra as rotas de gestão de usuário interno.
func RegisterUsuarioRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	h := NewHandler(c.CriarUsuarioUC, c.AtualizarUsuarioUC, c.AlterarSenhaUC, c.AtivarUsuarioUC, c.InativarUsuarioUC, c.BuscarUsuarioLogadoUC)

	autenticado := rg.Group("/usuarios", middleware.AuthenticationMiddleware(c.JWTAuth, c.RefreshTokensRepo, c.UsuarioStatusRepo, c.ClienteStatusRepo))

	admin := autenticado.Group("", middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelAdmin))
	admin.POST("", h.Criar)
	admin.PUT("/:id", h.Atualizar)
	admin.PATCH("/:id/ativar", h.Ativar)
	admin.PATCH("/:id/inativar", h.Inativar)

	self := autenticado.Group("", middleware.AuthorizationMiddleware(domainauth.TipoInterno))
	self.GET("/me", h.Me)
	self.PUT("/me/senha", h.AlterarSenha)
}
