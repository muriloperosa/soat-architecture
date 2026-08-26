package veiculo

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

type Handler struct {
	cadastrar         *appveiculo.CadastrarVeiculoUseCase
	atualizar         *appveiculo.AtualizarVeiculoUseCase
	ativar            *appveiculo.AtivarVeiculoUseCase
	inativar          *appveiculo.InativarVeiculoUseCase
	consultarPorID    *appveiculo.ConsultarVeiculoPorIDUseCase
	consultarPorPlaca *appveiculo.ConsultarVeiculoPorPlacaUseCase
}

func NewHandler(
	cadastrar *appveiculo.CadastrarVeiculoUseCase,
	atualizar *appveiculo.AtualizarVeiculoUseCase,
	ativar *appveiculo.AtivarVeiculoUseCase,
	inativar *appveiculo.InativarVeiculoUseCase,
	consultarPorID *appveiculo.ConsultarVeiculoPorIDUseCase,
	consultarPorPlaca *appveiculo.ConsultarVeiculoPorPlacaUseCase,
) *Handler {
	return &Handler{
		cadastrar:         cadastrar,
		atualizar:         atualizar,
		ativar:            ativar,
		inativar:          inativar,
		consultarPorID:    consultarPorID,
		consultarPorPlaca: consultarPorPlaca,
	}
}

// @Summary Cadastra veículo
// @Description Cadastra um veículo. Restrito a admin.
// @Tags Veiculos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CadastrarVeiculoRequest true "Dados do veículo"
// @Success 201 {object} VeiculoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Router /v1/veiculos [post]
func (h *Handler) Cadastrar(c *gin.Context) {
	criadoPor, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var req CadastrarVeiculoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.cadastrar.Executar(c.Request.Context(), toCadastrarInput(criadoPor, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toVeiculoResponse(out))
}

// @Summary Atualiza veículo
// @Description Atualiza dados cadastrais e quilometragem de um veículo. Quilometragem não pode regredir. Restrito a admin.
// @Tags Veiculos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do veículo"
// @Param request body AtualizarVeiculoRequest true "Dados atualizados"
// @Success 200 {object} VeiculoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/veiculos/{id} [put]
func (h *Handler) Atualizar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var req AtualizarVeiculoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.atualizar.Executar(c.Request.Context(), toAtualizarInput(id, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toVeiculoResponse(out))
}

// @Summary Ativa veículo
// @Description Reabilita um veículo. Restrito a admin.
// @Tags Veiculos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do veículo"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/veiculos/{id}/ativar [patch]
func (h *Handler) Ativar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.ativar.Executar(c.Request.Context(), id); err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Inativa veículo
// @Description Bloqueia um veículo. Restrito a admin.
// @Tags Veiculos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do veículo"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/veiculos/{id}/inativar [patch]
func (h *Handler) Inativar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}
	if _, err := h.inativar.Executar(c.Request.Context(), id); err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Consulta veículo por ID
// @Description Retorna os dados de um veículo pelo ID.
// @Tags Veiculos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do veículo"
// @Success 200 {object} VeiculoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/veiculos/{id} [get]
func (h *Handler) ConsultarPorID(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	out, err := h.consultarPorID.Executar(c.Request.Context(), id)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toVeiculoResponse(out))
}

// @Summary Consulta veículo por placa
// @Description Retorna os dados de um veículo pela placa.
// @Tags Veiculos
// @Produce json
// @Security BearerAuth
// @Param placa path string true "Placa do veículo"
// @Success 200 {object} VeiculoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/veiculos/placa/{placa} [get]
func (h *Handler) ConsultarPorPlaca(c *gin.Context) {
	placa := c.Param("placa")

	out, err := h.consultarPorPlaca.Executar(c.Request.Context(), placa)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toVeiculoResponse(out))
}
