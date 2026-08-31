package orcamento

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

type Handler struct {
	gerar               *app.GerarOrcamentoUseCase
	adicionarServico    *app.AdicionarServicoOrcamentoUseCase
	adicionarPeca       *app.AdicionarPecaOrcamentoUseCase
	removerServico      *app.RemoverServicoOrcamentoUseCase
	removerPeca         *app.RemoverPecaOrcamentoUseCase
	enviarParaAprovacao *app.EnviarOrcamentoParaAprovacaoUseCase
}

func NewHandler(
	gerar *app.GerarOrcamentoUseCase,
	adicionarServico *app.AdicionarServicoOrcamentoUseCase,
	adicionarPeca *app.AdicionarPecaOrcamentoUseCase,
	removerServico *app.RemoverServicoOrcamentoUseCase,
	removerPeca *app.RemoverPecaOrcamentoUseCase,
	enviarParaAprovacao *app.EnviarOrcamentoParaAprovacaoUseCase,
) *Handler {
	return &Handler{
		gerar:               gerar,
		adicionarServico:    adicionarServico,
		adicionarPeca:       adicionarPeca,
		removerServico:      removerServico,
		removerPeca:         removerPeca,
		enviarParaAprovacao: enviarParaAprovacao,
	}
}

// @Summary Gera o orçamento de uma Ordem de Serviço
// @Description Cria o orçamento de uma OS. Uma OS possui no máximo um orçamento. Restrito a mecânico ou administrador.
// @Tags Orçamentos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param request body GerarOrcamentoRequest true "Dados do orçamento"
// @Success 201 {object} OrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento [post]
func (h *Handler) Gerar(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	usuarioID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var request GerarOrcamentoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.gerar.Executar(c.Request.Context(), toGerarInput(ordemServicoID, usuarioID, request))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(output))
}

// @Summary Adiciona um serviço ao orçamento
// @Description Inclui um serviço do catálogo no orçamento da OS, copiando o valor e o tempo estimado vigentes. Restrito a mecânico ou administrador.
// @Tags Orçamentos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param request body AdicionarServicoOrcamentoRequest true "Serviço e quantidade"
// @Success 200 {object} OrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento/itens-servico [post]
func (h *Handler) AdicionarServico(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var request AdicionarServicoOrcamentoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.adicionarServico.Executar(c.Request.Context(), toAdicionarServicoInput(ordemServicoID, request))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Adiciona uma peça ao orçamento
// @Description Inclui uma peça do estoque no orçamento da OS, copiando a descrição e o valor vigentes. Restrito a mecânico ou administrador.
// @Tags Orçamentos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param request body AdicionarPecaOrcamentoRequest true "Peça e quantidade"
// @Success 200 {object} OrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento/itens-peca [post]
func (h *Handler) AdicionarPeca(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var request AdicionarPecaOrcamentoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.adicionarPeca.Executar(c.Request.Context(), toAdicionarPecaInput(ordemServicoID, request))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Remove um serviço do orçamento
// @Description Remove um item de serviço do orçamento da OS. Restrito a mecânico ou administrador.
// @Tags Orçamentos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param itemId path int true "ID do item de serviço"
// @Success 200 {object} OrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento/itens-servico/{itemId} [delete]
func (h *Handler) RemoverServico(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	itemServicoID, ok := httprequest.ParseUintParam(c, "itemId")
	if !ok {
		return
	}

	output, err := h.removerServico.Executar(c.Request.Context(), toRemoverServicoInput(ordemServicoID, itemServicoID))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Remove uma peça do orçamento
// @Description Remove um item de peça do orçamento da OS. Restrito a mecânico ou administrador.
// @Tags Orçamentos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param itemId path int true "ID do item de peça"
// @Success 200 {object} OrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento/itens-peca/{itemId} [delete]
func (h *Handler) RemoverPeca(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	itemPecaID, ok := httprequest.ParseUintParam(c, "itemId")
	if !ok {
		return
	}

	output, err := h.removerPeca.Executar(c.Request.Context(), toRemoverPecaInput(ordemServicoID, itemPecaID))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Envia o orçamento para aprovação
// @Description Move a OS de EM_DIAGNOSTICO ou REJEITADA para AGUARDANDO_APROVACAO, exige orçamento com itens e envia o orçamento ao cliente. Restrito a mecânico ou administrador.
// @Tags Orçamentos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Success 200 {object} OrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento/enviar-aprovacao [patch]
func (h *Handler) EnviarParaAprovacao(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	usuarioID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	output, err := h.enviarParaAprovacao.Executar(c.Request.Context(), app.EnviarOrcamentoParaAprovacaoInput{
		OrdemServicoID: ordemServicoID,
		UsuarioID:      usuarioID,
	})
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}
