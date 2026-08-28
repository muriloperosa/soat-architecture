package ordemservico

import (
	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

func RegisterOrdemServicoRoutes(rg *gin.RouterGroup, container *wiring.Container) {
	handler := NewHandler(container.AbrirOrdemServicoUC)

	ordensServico := rg.Group(
		"/ordens-servico",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
		middleware.AuthorizationMiddleware(domainauth.TipoInterno),
	)

	ordensServico.POST("", handler.Abrir)
}
