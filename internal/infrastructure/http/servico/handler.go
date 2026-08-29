package servico

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

// Handler expõe o CRUD do catálogo de serviços, restrito a usuário interno.
type Handler struct {
	criar       *appservico.CriarServicoUseCase
	atualizar   *appservico.AtualizarServicoUseCase
	listar      *appservico.ListarServicosUseCase
	buscar      *appservico.BuscarServicoUseCase
	ativar      *appservico.AtivarServicoUseCase
	inativar    *appservico.InativarServicoUseCase
	queryParser *httpquery.Parser
}

func NewHandler(
	criar *appservico.CriarServicoUseCase,
	atualizar *appservico.AtualizarServicoUseCase,
	listar *appservico.ListarServicosUseCase,
	buscar *appservico.BuscarServicoUseCase,
	ativar *appservico.AtivarServicoUseCase,
	inativar *appservico.InativarServicoUseCase,
	queryParser *httpquery.Parser,
) *Handler {
	return &Handler{
		criar:       criar,
		atualizar:   atualizar,
		listar:      listar,
		buscar:      buscar,
		ativar:      ativar,
		inativar:    inativar,
		queryParser: queryParser,
	}
}

// @Summary Cria serviço
// @Description Cadastra um item no catálogo de serviços da oficina. Restrito a usuário interno. criado_por sai do JWT.
// @Tags Servicos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CriarServicoRequest true "Dados do serviço"
// @Success 201 {object} ServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Router /v1/servicos [post]
func (h *Handler) Criar(c *gin.Context) {
	criadoPor, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var req CriarServicoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.criar.Executar(c.Request.Context(), toCriarInput(req, criadoPor))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toServicoResponse(out))
}

// @Summary Lista serviços
// @Description Lista serviços com paginação, ordenação e filtros diretos por campo. Texto usa LIKE, listas separadas por vírgula usam IN e booleanos usam igualdade. Duas datas ISO 8601 formam um intervalo. Use o sufixo _not para negar filtros.
// @Tags Servicos
// @Produce json
// @Security BearerAuth
// @Param offset query int false "Quantidade de registros ignorados" default(0) minimum(0)
// @Param limit query int false "Quantidade de registros retornados" default(20) minimum(1) maximum(100)
// @Param order query string false "Campo de ordenação" Enums(id,nome,descricao,preco_base,tempo_estimado_minutos,criado_por,ativo,data_cadastro,data_atualizacao) default(id)
// @Param direction query string false "Direção da ordenação" Enums(ASC,DESC) default(ASC)
// @Param id query string false "ID ou lista de IDs separada por vírgula" example(1,2,3)
// @Param nome query string false "Nome contendo o valor" example(óleo)
// @Param nome_not query string false "Nome que não deve conter o valor" example(Teste)
// @Param descricao query string false "Descrição contendo o valor"
// @Param preco_base query string false "Preço base ou lista de preços" example(150.50)
// @Param tempo_estimado_minutos query string false "Tempo estimado ou lista de tempos" example(30,60,90)
// @Param criado_por query string false "ID ou lista de IDs dos criadores"
// @Param ativo query bool false "Situação ativa do serviço"
// @Param data_cadastro query string false "Data ISO 8601 ou intervalo separado por vírgula" example(2026-08-20,2026-08-22)
// @Success 200 {object} ListarServicosResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/servicos [get]
func (h *Handler) Listar(c *gin.Context) {
	params, err := h.queryParser.Parse(c)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	out, err := h.listar.Executar(
		c.Request.Context(),
		params,
	)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(
		http.StatusOK,
		toListResponse(out),
	)
}

// @Summary Busca serviço
// @Description Retorna um serviço do catálogo pelo ID. Restrito a usuário interno.
// @Tags Servicos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do serviço"
// @Success 200 {object} ServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/servicos/{id} [get]
func (h *Handler) Buscar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	out, err := h.buscar.Executar(c.Request.Context(), id)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toServicoResponse(out))
}

// @Summary Atualiza serviço
// @Description Atualiza nome, descrição, preço base e tempo estimado. Não altera criado_por nem ativo. Restrito a usuário interno.
// @Tags Servicos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do serviço"
// @Param request body AtualizarServicoRequest true "Dados atualizados"
// @Success 200 {object} ServicoResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/servicos/{id} [put]
func (h *Handler) Atualizar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var req AtualizarServicoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.atualizar.Executar(c.Request.Context(), toAtualizarInput(id, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toServicoResponse(out))
}

// @Summary Ativa serviço
// @Description Reabilita um serviço para uso em Ordens de Serviço. Restrito a usuário interno.
// @Tags Servicos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do serviço"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/servicos/{id}/ativar [patch]
func (h *Handler) Ativar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.ativar.Executar(c.Request.Context(), id); err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Inativa serviço
// @Description Bloqueia um serviço de ser usado em novas Ordens de Serviço. Restrito a usuário interno.
// @Tags Servicos
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do serviço"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/servicos/{id}/inativar [patch]
func (h *Handler) Inativar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}
	if err := h.inativar.Executar(c.Request.Context(), id); err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}
