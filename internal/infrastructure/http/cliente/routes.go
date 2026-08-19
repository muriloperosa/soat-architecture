package cliente

import (
	"github.com/gin-gonic/gin"
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
	)

	clientes := rg.Group("/clientes")

	clientes.POST("", handler.Criar)

	clientes.GET("/documento/:documento", handler.BuscarPorDocumento)

	clientes.GET("/:id", handler.BuscarPorID)

	clientes.PUT("/:id", handler.Atualizar)

	clientes.PATCH("/:id/ativar", handler.Ativar)

	clientes.PATCH("/:id/inativar", handler.Inativar)

	clientes.PATCH("/:id/senha", handler.AlterarSenha)
}
