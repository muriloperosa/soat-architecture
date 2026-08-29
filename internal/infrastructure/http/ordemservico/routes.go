package ordemservico

import (
	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

func RegisterOrdemServicoRoutes(rg *gin.RouterGroup, container *wiring.Container) {
	handler := NewHandler(
		container.AbrirOrdemServicoUC,
		container.IniciarDiagnosticoUC,
		container.InformarDiagnosticoUC,
		container.IniciarExecucaoUC,
	)

	ordensServico := rg.Group(
		"/ordens-servico",
		middleware.AuthenticationMiddleware(container.JWTAuth, container.RefreshTokensRepo, container.UsuarioStatusRepo, container.ClienteStatusRepo),
		middleware.AuthorizationMiddleware(domainauth.TipoInterno),
	)

	ordensServico.POST("", handler.Abrir)

	ordensServicoExec := ordensServico.Group(
		"",
		middleware.AuthorizationMiddleware(domainauth.TipoInterno, shared.PapelMecanico, shared.PapelAdmin),
	)

	ordensServicoExec.PATCH("/:id/iniciar-diagnostico", handler.IniciarDiagnostico)
	ordensServicoExec.PUT("/:id/diagnostico", handler.InformarDiagnostico)
	ordensServicoExec.PATCH("/:id/iniciar-execucao", handler.IniciarExecucao)
}
