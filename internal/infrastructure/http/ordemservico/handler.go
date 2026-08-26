package ordemservico

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

type Handler struct {
	abrir *app.AbrirOrdemServicoUseCase
}

func NewHandler(abrir *app.AbrirOrdemServicoUseCase) *Handler {
	return &Handler{abrir: abrir}
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
