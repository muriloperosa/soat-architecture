package cliente

import (
	"net/http"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

// Handler expõe as operações HTTP de gestão de clientes.
type Handler struct {
	criar                 *app.CriarClienteUseCase
	atualizar             *app.AtualizarClienteUseCase
	consultarPorID        *app.ConsultarClientePorIDUseCase
	consultarPorDocumento *app.ConsultarClientePorDocumentoUseCase
	ativar                *app.AtivarClienteUseCase
	inativar              *app.InativarClienteUseCase
	alterarSenha          *app.AlterarSenhaClienteUseCase
	listar                *app.ListarClientesUseCase
	queryParser           *httpquery.Parser
}

func NewHandler(
	criar *app.CriarClienteUseCase,
	atualizar *app.AtualizarClienteUseCase,
	consultarPorID *app.ConsultarClientePorIDUseCase,
	consultarPorDocumento *app.ConsultarClientePorDocumentoUseCase,
	ativar *app.AtivarClienteUseCase,
	inativar *app.InativarClienteUseCase,
	alterarSenha *app.AlterarSenhaClienteUseCase,
	listar *app.ListarClientesUseCase,
	queryParser *httpquery.Parser,
) *Handler {
	return &Handler{
		criar:                 criar,
		atualizar:             atualizar,
		consultarPorID:        consultarPorID,
		consultarPorDocumento: consultarPorDocumento,
		ativar:                ativar,
		inativar:              inativar,
		alterarSenha:          alterarSenha,
		listar:                listar,
		queryParser:           queryParser,
	}
}

// @Summary Cria cliente
// @Description Cria um cliente para atendimento da oficina.
// @Tags Clientes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CriarClienteRequest true "Dados do cliente"
// @Success 201 {object} ClienteResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Router /v1/clientes [post]
func (h *Handler) Criar(c *gin.Context) {
	criadoPor, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var req CriarClienteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.criar.Executar(c.Request.Context(), toCriarInput(criadoPor, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toResponse(output))
}

// @Summary Lista clientes
// @Description Lista clientes com paginação, ordenação e filtros diretos por campo. Texto usa LIKE, listas separadas por vírgula usam IN, booleanos usam igualdade e duas datas ISO 8601 formam um intervalo. Use o sufixo _not para negar filtros textuais.
// @Tags Clientes
// @Produce json
// @Security BearerAuth
// @Param page query int false "Número da página" default(1) minimum(1)
// @Param order query string false "Campo de ordenação" Enums(id,documento,tipo,nome,email,telefone,ativo,requer_alterar_senha,criado_por,data_cadastro,data_atualizacao) default(id)
// @Param direction query string false "Direção da ordenação" Enums(ASC,DESC) default(ASC)
// @Param id query string false "ID ou lista de IDs separada por vírgula" example(1,2,3)
// @Param documento query string false "Documento contendo o valor"
// @Param tipo query string false "Tipo de pessoa contendo o valor" Enums(PF,PJ)
// @Param nome query string false "Nome contendo o valor" example(Maria)
// @Param nome_not query string false "Nome que não deve conter o valor" example(Teste)
// @Param email query string false "E-mail contendo o valor"
// @Param telefone query string false "Telefone contendo o valor"
// @Param ativo query bool false "Situação ativa do cliente"
// @Param criado_por query string false "ID ou lista de IDs dos criadores"
// @Param data_cadastro query string false "Data ISO 8601 ou intervalo separado por vírgula" example(2026-08-20,2026-08-22)
// @Success 200 {object} ListarClientesResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/clientes [get]
func (h *Handler) Listar(c *gin.Context) {
	params, err := h.queryParser.Parse(c)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	output, err := h.listar.Executar(c.Request.Context(), toListarInput(params))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toListResponse(output))
}

// @Summary Atualiza cliente
// @Description Atualiza os dados cadastrais de um cliente.
// @Tags Clientes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do cliente"
// @Param request body AtualizarClienteRequest true "Dados atualizados"
// @Success 200 {object} ClienteResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/clientes/{id} [put]
func (h *Handler) Atualizar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var req AtualizarClienteRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	output, err := h.atualizar.Executar(c.Request.Context(), toAtualizarInput(id, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Busca cliente por ID
// @Description Retorna os dados de um cliente pelo identificador.
// @Tags Clientes
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do cliente"
// @Success 200 {object} ClienteResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/clientes/{id} [get]
func (h *Handler) BuscarPorID(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	output, err := h.consultarPorID.Executar(c.Request.Context(), id)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Busca cliente por documento
// @Description Retorna os dados de um cliente pelo CPF ou CNPJ.
// @Tags Clientes
// @Produce json
// @Security BearerAuth
// @Param documento path string true "CPF ou CNPJ do cliente"
// @Success 200 {object} ClienteResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/clientes/documento/{documento} [get]
func (h *Handler) BuscarPorDocumento(c *gin.Context) {
	documento := c.Param("documento")

	output, err := h.consultarPorDocumento.Executar(c.Request.Context(), documento)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toResponse(output))
}

// @Summary Ativa cliente
// @Description Reabilita o cliente para login e uso do sistema.
// @Tags Clientes
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do cliente"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/clientes/{id}/ativar [patch]
func (h *Handler) Ativar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	_, err := h.ativar.Executar(c.Request.Context(), id)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Inativa cliente
// @Description Bloqueia o cliente para login e uso do sistema.
// @Tags Clientes
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do cliente"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/clientes/{id}/inativar [patch]
func (h *Handler) Inativar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	_, err := h.inativar.Executar(c.Request.Context(), id)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// @Summary Altera senha do cliente
// @Description Altera a senha do cliente e remove a exigência de troca de senha.
// @Tags Clientes
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AlterarSenhaRequest true "Nova senha"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/clientes/me/senha [put]
func (h *Handler) AlterarSenha(c *gin.Context) {
	id, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	var req AlterarSenhaRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	_, err := h.alterarSenha.Executar(c.Request.Context(), req.toInput(id))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
