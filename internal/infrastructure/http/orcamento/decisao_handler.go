package orcamento

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

type DecisaoHandler struct {
	aprovar  *app.AprovarOrcamentoUseCase
	rejeitar *app.RejeitarOrcamentoUseCase
}

func NewDecisaoHandler(
	aprovar *app.AprovarOrcamentoUseCase,
	rejeitar *app.RejeitarOrcamentoUseCase,
) *DecisaoHandler {
	return &DecisaoHandler{aprovar: aprovar, rejeitar: rejeitar}
}

// @Summary Aprova o orçamento da própria Ordem de Serviço
// @Description Cliente autenticado aprova o orçamento somente quando a OS está AGUARDANDO_APROVACAO e pertence a ele.
// @Tags Orçamentos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Success 200 {object} FluxoOrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento/aprovar [patch]
func (h *DecisaoHandler) Aprovar(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	clienteID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	output, err := h.aprovar.Executar(c.Request.Context(), app.AprovarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ClienteID:      clienteID,
	})
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toFluxoResponse(output))
}

// @Summary Rejeita o orçamento da própria Ordem de Serviço
// @Description Cliente autenticado rejeita o orçamento somente quando a OS está AGUARDANDO_APROVACAO e pertence a ele. O motivo é registrado no histórico da OS.
// @Tags Orçamentos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param request body RejeitarOrcamentoRequest true "Motivo da rejeição"
// @Success 200 {object} FluxoOrcamentoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/orcamento/rejeitar [patch]
func (h *DecisaoHandler) Rejeitar(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	clienteID, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var request RejeitarOrcamentoRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.rejeitar.Executar(c.Request.Context(), app.RejeitarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ClienteID:      clienteID,
		Motivo:         request.Motivo,
	})
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toFluxoResponse(output))
}
