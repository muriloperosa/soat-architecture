package ordemservico

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

type Handler struct {
	abrir               *app.AbrirOrdemServicoUseCase
	iniciarDiagnostico  *app.IniciarDiagnosticoUseCase
	informarDiagnostico *app.InformarDiagnosticoUseCase
	iniciarExecucao     *app.IniciarExecucaoUseCase
}

func NewHandler(
	abrir *app.AbrirOrdemServicoUseCase,
	iniciarDiagnostico *app.IniciarDiagnosticoUseCase,
	informarDiagnostico *app.InformarDiagnosticoUseCase,
	iniciarExecucao *app.IniciarExecucaoUseCase,
) *Handler {
	return &Handler{
		abrir:               abrir,
		iniciarDiagnostico:  iniciarDiagnostico,
		informarDiagnostico: informarDiagnostico,
		iniciarExecucao:     iniciarExecucao,
	}
}

// @Summary Abre uma Ordem de Serviço
// @Description Cria uma Ordem de Serviço no status RECEBIDA e registra o histórico inicial. Restrito a usuário interno autenticado.
// @Tags Ordens de Serviço
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AbrirOrdemServicoRequest true "Dados da Ordem de Serviço"
// @Success 201 {object} OrdemServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico [post]
func (h *Handler) Abrir(c *gin.Context) {
	usuarioID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var request AbrirOrdemServicoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.abrir.Executar(c.Request.Context(), toInput(usuarioID, request))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(output))
}

// @Summary Inicia o diagnóstico de uma Ordem de Serviço
// @Description Altera uma Ordem de Serviço RECEBIDA para EM_DIAGNOSTICO e registra o histórico. Restrito a mecânico ou administrador.
// @Tags Ordens de Serviço
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Success 200 {object} OrdemServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/iniciar-diagnostico [patch]
func (h *Handler) IniciarDiagnostico(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	usuarioID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	output, err := h.iniciarDiagnostico.Executar(
		c.Request.Context(),
		toIniciarDiagnosticoInput(id, usuarioID),
	)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Informa o diagnóstico de uma Ordem de Serviço
// @Description Registra o diagnóstico realizado enquanto a Ordem de Serviço está EM_DIAGNOSTICO. Restrito a mecânico ou administrador.
// @Tags Ordens de Serviço
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param request body InformarDiagnosticoRequest true "Diagnóstico realizado"
// @Success 200 {object} OrdemServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/diagnostico [put]
func (h *Handler) InformarDiagnostico(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var request InformarDiagnosticoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.informarDiagnostico.Executar(
		c.Request.Context(),
		toInformarDiagnosticoInput(id, request),
	)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Inicia a execução de uma Ordem de Serviço
// @Description Altera uma Ordem de Serviço APROVADA para EM_EXECUCAO e registra o histórico. Restrito a usuário interno autenticado.
// @Tags Ordens de Serviço
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Success 200 {object} OrdemServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/iniciar-execucao [patch]
func (h *Handler) IniciarExecucao(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	usuarioID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	output, err := h.iniciarExecucao.Executar(
		c.Request.Context(),
		toIniciarExecucaoInput(id, usuarioID),
	)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}
