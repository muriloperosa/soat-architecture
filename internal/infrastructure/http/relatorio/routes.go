package relatorio

import (
	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

func RegisterRelatorioRoutes(rg *gin.RouterGroup, container *wiring.Container) {
	handler := NewHandler(container.ConsultarTransicaoStatusUC)

	relatorios := rg.Group(
		"/relatorios",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
		middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelAdmin),
	)

	relatorios.GET("/ordens-servico/transicao-status", handler.ConsultarTransicaoStatus)
}
