package ordemservico

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

type Handler struct {
	consultarPorID      *app.ConsultarOrdemServicoPorIDUseCase
	consultarPorNumero  *app.ConsultarOrdemServicoPorNumeroUseCase
	listar              *app.ListarOrdensServicoUseCase
	queryParser         *httpquery.Parser
	abrir               *app.AbrirOrdemServicoUseCase
	iniciarDiagnostico  *app.IniciarDiagnosticoUseCase
	informarDiagnostico *app.InformarDiagnosticoUseCase
	iniciarExecucao     *app.IniciarExecucaoUseCase
	finalizar           *app.FinalizarOrdemServicoUseCase
	entregar            *app.EntregarOrdemServicoUseCase
}

func NewHandler(
	abrir *app.AbrirOrdemServicoUseCase,
	iniciarDiagnostico *app.IniciarDiagnosticoUseCase,
	informarDiagnostico *app.InformarDiagnosticoUseCase,
	iniciarExecucao *app.IniciarExecucaoUseCase,
	finalizar *app.FinalizarOrdemServicoUseCase,
	entregar *app.EntregarOrdemServicoUseCase,
	consultarPorID *app.ConsultarOrdemServicoPorIDUseCase,
	consultarPorNumero *app.ConsultarOrdemServicoPorNumeroUseCase,
	listar *app.ListarOrdensServicoUseCase,
	queryParser *httpquery.Parser,
) *Handler {
	return &Handler{
		consultarPorID:      consultarPorID,
		consultarPorNumero:  consultarPorNumero,
		listar:              listar,
		queryParser:         queryParser,
		abrir:               abrir,
		iniciarDiagnostico:  iniciarDiagnostico,
		informarDiagnostico: informarDiagnostico,
		iniciarExecucao:     iniciarExecucao,
		finalizar:           finalizar,
		entregar:            entregar,
	}
}

// @Summary Lista Ordens de Serviço
// @Description Lista Ordens de Serviço com paginação, ordenação e filtros diretos por campo. Usuários do tipo cliente visualizam somente as próprias Ordens de Serviço.
// @Tags Ordens de Serviço
// @Produce json
// @Security BearerAuth
// @Param page query int false "Número da página" default(1) minimum(1)
// @Param order query string false "Campo de ordenação" Enums(id,numero,cliente_id,veiculo_id,quilometragem_entrada,status,criado_por,data_cadastro,data_atualizacao) default(id)
// @Param direction query string false "Direção da ordenação" Enums(ASC,DESC) default(ASC)
// @Param status query string false "Status ou lista de status separados por vírgula"
// @Param cliente_id query string false "ID ou lista de IDs de clientes separados por vírgula"
// @Param veiculo_id query string false "ID ou lista de IDs de veículos separados por vírgula"
// @Success 200 {object} ListarOrdensServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico [get]
func (h *Handler) Listar(c *gin.Context) {
	solicitanteID, tipoSolicitante, ok := middleware.Subject(c)
	if !ok {
		return
	}

	params, err := h.queryParser.Parse(c)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	output, err := h.listar.Executar(
		c.Request.Context(),
		toListarInput(params, solicitanteID, tipoSolicitante),
	)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toListResponse(output))
}

// @Summary Busca uma Ordem de Serviço por ID
// @Description Usuários internos podem consultar qualquer OS. Clientes podem consultar somente Ordens de Serviço vinculadas ao próprio cadastro.
// @Tags Ordens de Serviço
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Success 200 {object} OrdemServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id} [get]
func (h *Handler) BuscarPorID(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	solicitanteID, tipoSolicitante, ok := middleware.Subject(c)
	if !ok {
		return
	}

	output, err := h.consultarPorID.Executar(c.Request.Context(), app.ConsultarOrdemServicoPorIDInput{
		ID:              id,
		SolicitanteID:   solicitanteID,
		TipoSolicitante: tipoSolicitante,
	})
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Busca uma Ordem de Serviço por número
// @Description Usuários internos podem consultar qualquer OS. Clientes podem consultar somente Ordens de Serviço vinculadas ao próprio cadastro.
// @Tags Ordens de Serviço
// @Produce json
// @Security BearerAuth
// @Param numero path string true "Número da Ordem de Serviço"
// @Success 200 {object} OrdemServicoResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/numero/{numero} [get]
func (h *Handler) BuscarPorNumero(c *gin.Context) {
	solicitanteID, tipoSolicitante, ok := middleware.Subject(c)
	if !ok {
		return
	}

	output, err := h.consultarPorNumero.Executar(c.Request.Context(), app.ConsultarOrdemServicoPorNumeroInput{
		Numero:          c.Param("numero"),
		SolicitanteID:   solicitanteID,
		TipoSolicitante: tipoSolicitante,
	})
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
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
// @Description Altera uma Ordem de Serviço APROVADA para EM_EXECUCAO e registra o histórico. Restrito a mecânico ou administrador.
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

// @Summary Finaliza uma Ordem de Serviço
// @Description Consome as reservas de peças (estoque físico), remove as reservas e altera a Ordem de Serviço EM_EXECUCAO para FINALIZADA, registrando o histórico. Tudo ocorre na mesma transação. Restrito a mecânico ou administrador.
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
// @Router /v1/ordens-servico/{id}/finalizar [patch]
func (h *Handler) Finalizar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	usuarioID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	output, err := h.finalizar.Executar(
		c.Request.Context(),
		toFinalizarInput(id, usuarioID),
	)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Entrega uma Ordem de Serviço
// @Description Altera uma Ordem de Serviço FINALIZADA para ENTREGUE e registra o histórico. Encerra o ciclo da OS. Restrito a usuário interno autenticado.
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
// @Router /v1/ordens-servico/{id}/entregar [patch]
func (h *Handler) Entregar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	usuarioID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	output, err := h.entregar.Executar(
		c.Request.Context(),
		toEntregarInput(id, usuarioID),
	)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}
