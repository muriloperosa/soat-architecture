package peca

import (
	"net/http"

	"github.com/gin-gonic/gin"

	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
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
	reporEstoque   *apppeca.ReporEstoqueUseCase
}

func NewHandler(
	cadastrar *apppeca.CadastrarPecaUseCase,
	atualizar *apppeca.AtualizarPecaUseCase,
	ativar *apppeca.AtivarPecaUseCase,
	inativar *apppeca.InativarPecaUseCase,
	consultarPorID *apppeca.ConsultarPecaPorIDUseCase,
	reporEstoque *apppeca.ReporEstoqueUseCase,
) *Handler {
	return &Handler{
		cadastrar:      cadastrar,
		atualizar:      atualizar,
		ativar:         ativar,
		inativar:       inativar,
		consultarPorID: consultarPorID,
		reporEstoque:   reporEstoque,
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
