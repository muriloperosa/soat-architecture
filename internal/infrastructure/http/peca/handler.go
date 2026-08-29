package peca

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

// Handler expõe a gestão de peças do estoque: CRUD e reposição de estoque,
// restritos a admin.
type Handler struct {
	cadastrar      *apppeca.CadastrarPecaUseCase
	atualizar      *apppeca.AtualizarPecaUseCase
	ativar         *apppeca.AtivarPecaUseCase
	inativar       *apppeca.InativarPecaUseCase
	consultarPorID *apppeca.ConsultarPecaPorIDUseCase
	listar         *apppeca.ListarPecasUseCase
	reporEstoque   *apppeca.ReporEstoqueUseCase
	queryParser    *httpquery.Parser
}

func NewHandler(
	cadastrar *apppeca.CadastrarPecaUseCase,
	atualizar *apppeca.AtualizarPecaUseCase,
	ativar *apppeca.AtivarPecaUseCase,
	inativar *apppeca.InativarPecaUseCase,
	consultarPorID *apppeca.ConsultarPecaPorIDUseCase,
	listar *apppeca.ListarPecasUseCase,
	reporEstoque *apppeca.ReporEstoqueUseCase,
	queryParser *httpquery.Parser,
) *Handler {
	return &Handler{
		cadastrar:      cadastrar,
		atualizar:      atualizar,
		ativar:         ativar,
		inativar:       inativar,
		consultarPorID: consultarPorID,
		listar:         listar,
		reporEstoque:   reporEstoque,
		queryParser:    queryParser,
	}
}

// @Summary Cadastra peça
// @Description Cadastra uma peça de estoque. Restrito a admin.
// @Tags Pecas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CadastrarPecaRequest true "Dados da peça"
// @Success 201 {object} PecaResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Router /v1/pecas [post]
func (h *Handler) Cadastrar(c *gin.Context) {
	criadoPor, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var req CadastrarPecaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.cadastrar.Executar(c.Request.Context(), toCadastrarInput(criadoPor, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toPecaResponse(out))
}

// @Summary Atualiza peça
// @Description Atualiza dados cadastrais e estoque mínimo de uma peça. Restrito a admin.
// @Tags Pecas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da peça"
// @Param request body AtualizarPecaRequest true "Dados atualizados"
// @Success 200 {object} PecaResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/pecas/{id} [put]
func (h *Handler) Atualizar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var req AtualizarPecaRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.atualizar.Executar(c.Request.Context(), toAtualizarInput(id, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPecaResponse(out))
}

// @Summary Ativa peça
// @Description Reabilita uma peça pra uso em Ordens de Serviço. Restrito a admin.
// @Tags Pecas
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da peça"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/pecas/{id}/ativar [patch]
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

// @Summary Inativa peça
// @Description Bloqueia uma peça de ser usada em novas Ordens de Serviço. Restrito a admin.
// @Tags Pecas
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da peça"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/pecas/{id}/inativar [patch]
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

// @Summary Consulta peça por ID
// @Description Retorna os dados de uma peça pelo ID.
// @Tags Pecas
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da peça"
// @Success 200 {object} PecaResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/pecas/{id} [get]
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

	c.JSON(http.StatusOK, toPecaResponse(out))
}

// @Summary Repõe estoque de peça
// @Description Adiciona quantidade ao estoque físico de uma peça (ex. entrada de fornecedor). Restrito a admin.
// @Tags Pecas
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID da peça"
// @Param request body ReporEstoqueRequest true "Quantidade a repor"
// @Success 200 {object} PecaResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/pecas/{id}/repor-estoque [patch]
func (h *Handler) ReporEstoque(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var req ReporEstoqueRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.reporEstoque.Executar(c.Request.Context(), toReporEstoqueInput(id, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toPecaResponse(out))
}

// @Summary Lista peças
// @Description Lista peças do estoque com paginação, ordenação e filtros.
// @Tags Pecas
// @Produce json
// @Security BearerAuth
// @Param offset query int false "Quantidade de registros a ignorar" default(0) minimum(0)
// @Param limit query int false "Quantidade máxima de registros" default(20) minimum(1) maximum(100)
// @Param order query string false "Campo para ordenação" Enums(id,codigo,nome,marca,descricao,preco,quantidade_em_estoque,estoque_minimo,criado_por,ativo,data_cadastro,data_atualizacao) default(id)
// @Param direction query string false "Direção da ordenação" Enums(ASC,DESC) default(ASC)
// @Param id query int false "Filtra pelo ID"
// @Param codigo query string false "Filtra pelo código"
// @Param nome query string false "Filtra pelo nome"
// @Param marca query string false "Filtra pela marca"
// @Param descricao query string false "Filtra pela descrição"
// @Param preco query number false "Filtra pelo preço"
// @Param quantidade_em_estoque query int false "Filtra pela quantidade em estoque"
// @Param estoque_minimo query int false "Filtra pelo estoque mínimo"
// @Param criado_por query int false "Filtra pelo usuário que cadastrou"
// @Param ativo query bool false "Filtra pelo status ativo"
// @Param data_cadastro query string false "Filtra pela data de cadastro"
// @Success 200 {object} ListarPecasResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/pecas [get]
func (h *Handler) Listar(c *gin.Context) {
	params, err := h.queryParser.Parse(c)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	out, err := h.listar.Executar(c.Request.Context(), params)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toListResponse(out))
}
