package cliente

import (
	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// RegisterClienteRoutes registra as rotas de gestão de clientes.
func RegisterClienteRoutes(
	rg *gin.RouterGroup,
	container *wiring.Container,
) {
	handler := NewHandler(
		container.CriarClienteUseCase,
		container.AtualizarClienteUseCase,
		container.ConsultarClientePorIDUseCase,
		container.ConsultarClientePorDocumentoUseCase,
		container.AtivarClienteUseCase,
		container.InativarClienteUseCase,
		container.AlterarSenhaClienteUseCase,
		container.ListarClientesUseCase,
		httpquery.NewParser(),
	)

	clientes := rg.Group(
		"/clientes",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
		middleware.AuthorizationMiddleware(domainauth.TipoInterno),
	)

	clientes.POST("", handler.Criar)
	clientes.GET("", handler.Listar)

	clientes.GET("/documento/:documento", handler.BuscarPorDocumento)

	clientes.GET("/:id", handler.BuscarPorID)

	clientes.PUT("/:id", handler.Atualizar)

	clientes.PATCH("/:id/ativar", handler.Ativar)

	clientes.PATCH("/:id/inativar", handler.Inativar)

	self := rg.Group(
		"/clientes",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
		middleware.AuthorizationMiddleware(domainauth.TipoCliente),
	)
	self.PUT("/me/senha", handler.AlterarSenha)
}
