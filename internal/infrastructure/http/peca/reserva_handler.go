package peca

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/peca"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
)

type ReservaHandler struct {
	reservar          *app.ReservarPecaUseCase
	alterarQuantidade *app.AlterarQuantidadeReservaPecaUseCase
}

func NewReservaHandler(
	reservar *app.ReservarPecaUseCase,
	alterarQuantidade *app.AlterarQuantidadeReservaPecaUseCase,
) *ReservaHandler {
	return &ReservaHandler{reservar: reservar, alterarQuantidade: alterarQuantidade}
}

// @Summary Reserva uma peça para uma Ordem de Serviço
// @Description Cria ou incrementa a reserva da peça para a OS, respeitando estoque mínimo e protegendo concorrência com transação e FOR UPDATE.
// @Tags Reservas de Peças
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param request body ReservarPecaRequest true "Peça e quantidade a reservar"
// @Success 200 {object} ReservaPecaResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/reservas-pecas [post]
func (h *ReservaHandler) Reservar(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var request ReservarPecaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.reservar.Executar(c.Request.Context(), app.ReservarPecaInput{
		PecaID:         request.PecaID,
		OrdemServicoID: ordemServicoID,
		Quantidade:     request.Quantidade,
	})
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toReservaResponse(output))
}

// @Summary Altera a quantidade total reservada de uma peça na OS
// @Description Atualiza a quantidade total da reserva existente, recalculando disponibilidade dentro de transação com lock da peça.
// @Tags Reservas de Peças
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da Ordem de Serviço"
// @Param pecaId path int true "ID da peça"
// @Param request body AlterarQuantidadeReservaPecaRequest true "Nova quantidade total reservada"
// @Success 200 {object} ReservaPecaResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/ordens-servico/{id}/reservas-pecas/{pecaId} [put]
func (h *ReservaHandler) AlterarQuantidade(c *gin.Context) {
	ordemServicoID, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}
	pecaID, ok := httprequest.ParseUintParam(c, "pecaId")
	if !ok {
		return
	}

	var request AlterarQuantidadeReservaPecaRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.alterarQuantidade.Executar(c.Request.Context(), app.AlterarQuantidadeReservaPecaInput{
		PecaID:         pecaID,
		OrdemServicoID: ordemServicoID,
		Quantidade:     request.Quantidade,
	})
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toReservaResponse(output))
}

func toReservaResponse(output app.ReservaPecaOutput) ReservaPecaResponse {
	return ReservaPecaResponse{
		ID:             output.ID,
		OrdemServicoID: output.OrdemServicoID,
		PecaID:         output.PecaID,
		Quantidade:     output.Quantidade,
	}
}
