package relatorio

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/relatorio"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
)

type Handler struct {
	consultarTransicaoStatus *app.ConsultarTransicaoStatusUseCase
}

func NewHandler(consultarTransicaoStatus *app.ConsultarTransicaoStatusUseCase) *Handler {
	return &Handler{consultarTransicaoStatus: consultarTransicaoStatus}
}

// @Summary Relatório de transição de status de Ordens de Serviço
// @Description Calcula quantas OS fizeram a transição from_status->to_status no período informado e a duração média/mínima/máxima. Restrito a administrador.
// @Tags Relatórios
// @Produce json
// @Security BearerAuth
// @Param start_date query string true "Data inicial (YYYY-MM-DD)"
// @Param final_date query string true "Data final (YYYY-MM-DD)"
// @Param from_status query string true "Status de origem"
// @Param to_status query string true "Status de destino"
// @Param unit query string false "Unidade de tempo da resposta: h (horas - padrão), m (minutos) ou s (segundos)"
// @Success 200 {object} TransicaoStatusResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/relatorios/ordens-servico/transicao-status [get]
func (h *Handler) ConsultarTransicaoStatus(c *gin.Context) {
	var request ConsultarTransicaoStatusRequest
	if err := c.ShouldBindQuery(&request); err != nil {
		httperror.RespondValidationError(c, "Query params inválidos.")
		return
	}

	unidade, ok := normalizarUnidade(request.Unit)
	if !ok {
		httperror.RespondValidationError(c, "unit inválido, valores aceitos: h, m, s.")
		return
	}

	dataInicio, ok := httprequest.ParseDateQueryParam(c, "start_date")
	if !ok {
		return
	}

	dataFim, ok := httprequest.ParseDateQueryParam(c, "final_date")
	if !ok {
		return
	}

	output, err := h.consultarTransicaoStatus.Executar(c.Request.Context(), toInput(request, dataInicio, dataFim))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output, unidade))
}

func normalizarUnidade(unit string) (string, bool) {
	if unit == "" {
		return unidadeHoras, true
	}
	switch unit {
	case unidadeHoras, unidadeMinutos, unidadeSegundos:
		return unit, true
	default:
		return "", false
	}
}
