package veiculo

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
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
	listar            *appveiculo.ListarVeiculosUseCase
	queryParser       *httpquery.Parser
}

func NewHandler(
	cadastrar *appveiculo.CadastrarVeiculoUseCase,
	atualizar *appveiculo.AtualizarVeiculoUseCase,
	ativar *appveiculo.AtivarVeiculoUseCase,
	inativar *appveiculo.InativarVeiculoUseCase,
	consultarPorID *appveiculo.ConsultarVeiculoPorIDUseCase,
	consultarPorPlaca *appveiculo.ConsultarVeiculoPorPlacaUseCase,
	listar *appveiculo.ListarVeiculosUseCase,
	queryParser *httpquery.Parser,
) *Handler {
	return &Handler{
		cadastrar:         cadastrar,
		atualizar:         atualizar,
		ativar:            ativar,
		inativar:          inativar,
		consultarPorID:    consultarPorID,
		consultarPorPlaca: consultarPorPlaca,
		listar:            listar,
		queryParser:       queryParser,
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

// @Summary Lista veículos
// @Description Lista veículos com paginação, ordenação e filtros diretos por campo. Campos de texto usam LIKE, listas separadas por vírgula usam IN e booleanos usam igualdade. Duas datas ISO 8601 formam um intervalo. Use o sufixo _not para negar filtros.
// @Tags Veiculos
// @Produce json
// @Security BearerAuth
// @Param page query int false "Número da página" default(1) minimum(1)
// @Param order query string false "Campo de ordenação" Enums(id,placa,marca,modelo,quilometragem_atual,ano,cor,criado_por,ativo,data_cadastro,data_atualizacao) default(id)
// @Param direction query string false "Direção da ordenação" Enums(ASC,DESC) default(ASC)
// @Param id query string false "ID ou lista de IDs separada por vírgula" example(1,2,3)
// @Param placa query string false "Placa contendo o valor" example(ABC1D23)
// @Param placa_not query string false "Placa que não deve conter o valor" example(ABC)
// @Param marca query string false "Marca contendo o valor" example(Fiat)
// @Param marca_not query string false "Marca que não deve conter o valor" example(Ford)
// @Param modelo query string false "Modelo contendo o valor" example(Uno)
// @Param modelo_not query string false "Modelo que não deve conter o valor" example(Palio)
// @Param quilometragem_atual query string false "Quilometragem atual ou lista de valores" example(15000,30000)
// @Param ano query string false "Ano ou lista de anos" example(2020,2021,2022)
// @Param cor query string false "Cor contendo o valor" example(Prata)
// @Param cor_not query string false "Cor que não deve conter o valor" example(Preto)
// @Param criado_por query string false "ID ou lista de IDs dos usuários que cadastraram o veículo"
// @Param ativo query bool false "Situação ativa do veículo"
// @Param data_cadastro query string false "Data ISO 8601 ou intervalo separado por vírgula" example(2026-08-20,2026-08-22)
// @Success 200 {object} ListarVeiculosResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/veiculos [get]
func (h *Handler) Listar(c *gin.Context) {
	params, err := h.queryParser.Parse(c)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	out, err := h.listar.Executar(c.Request.Context(), toListarInput(params))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toListResponse(out))
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
