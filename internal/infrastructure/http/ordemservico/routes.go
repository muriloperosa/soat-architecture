package ordemservico

import (
	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

func RegisterOrdemServicoRoutes(rg *gin.RouterGroup, container *wiring.Container) {
	handler := NewHandler(
		container.AbrirOrdemServicoUC,
		container.IniciarDiagnosticoUC,
		container.InformarDiagnosticoUC,
		container.IniciarExecucaoUC,
		container.EntregarOrdemServicoUC,
		container.ConsultarOrdemServicoPorIDUC,
		container.ConsultarOrdemServicoPorNumeroUC,
		container.ListarOrdensServicoUC,
		httpquery.NewParser(),
	)

	ordensServico := rg.Group(
		"/ordens-servico",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
	)

	// Consultas são acessíveis por usuário interno e cliente autenticado.
	// Para clientes, a aplicação restringe a consulta às próprias Ordens de Serviço.
	ordensServico.GET("", handler.Listar)
	ordensServico.GET("/numero/:numero", handler.BuscarPorNumero)
	ordensServico.GET("/:id", handler.BuscarPorID)

	ordensServicoInternas := ordensServico.Group(
		"",
		middleware.AuthorizationMiddleware(domainauth.TipoInterno),
	)

	ordensServicoInternas.POST("", handler.Abrir)
	ordensServicoInternas.PATCH("/:id/entregar", handler.Entregar)

	ordensServicoExec := ordensServicoInternas.Group(
		"",
		middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelMecanico, shared.PapelAdmin),
	)

	ordensServicoExec.PATCH("/:id/iniciar-diagnostico", handler.IniciarDiagnostico)
	ordensServicoExec.PUT("/:id/diagnostico", handler.InformarDiagnostico)
	ordensServicoExec.PATCH("/:id/iniciar-execucao", handler.IniciarExecucao)
}
