package servico

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

// Handler expõe o CRUD do catálogo de serviços, restrito a usuário interno.
type Handler struct {
	criar     *appservico.CriarServicoUseCase
	atualizar *appservico.AtualizarServicoUseCase
	listar    *appservico.ListarServicosUseCase
	buscar    *appservico.BuscarServicoUseCase
	ativar    *appservico.AtivarServicoUseCase
	inativar  *appservico.InativarServicoUseCase
}

func NewHandler(
	criar *appservico.CriarServicoUseCase,
	atualizar *appservico.AtualizarServicoUseCase,
	listar *appservico.ListarServicosUseCase,
	buscar *appservico.BuscarServicoUseCase,
	ativar *appservico.AtivarServicoUseCase,
	inativar *appservico.InativarServicoUseCase,
) *Handler {
	return &Handler{
		criar:     criar,
		atualizar: atualizar,
		listar:    listar,
		buscar:    buscar,
		ativar:    ativar,
		inativar:  inativar,
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
// @Description Lista o catálogo completo de serviços. Restrito a usuário interno.
// @Tags Servicos
// @Produce json
// @Security BearerAuth
// @Success 200 {array} ServicoResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Router /v1/servicos [get]
func (h *Handler) Listar(c *gin.Context) {
	out, err := h.listar.Executar(c.Request.Context())
	if err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toServicoResponseList(out))
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
