package usuario

import (
	"net/http"

	"github.com/gin-gonic/gin"

	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
)

// Handler expõe a gestão de usuários internos: CRUD restrito a admin
// (Criar/Atualizar/Ativar/Inativar) e autoatendimento pro próprio usuário
// logado (Me/AlterarSenha).
type Handler struct {
	criar        *appusuario.CriarUsuarioUseCase
	atualizar    *appusuario.AtualizarUsuarioUseCase
	alterarSenha *appusuario.AlterarSenhaUseCase
	ativar       *appusuario.AtivarUsuarioUseCase
	inativar     *appusuario.InativarUsuarioUseCase
	buscarLogado *appusuario.BuscarUsuarioLogadoUseCase
}

func NewHandler(
	criar *appusuario.CriarUsuarioUseCase,
	atualizar *appusuario.AtualizarUsuarioUseCase,
	alterarSenha *appusuario.AlterarSenhaUseCase,
	ativar *appusuario.AtivarUsuarioUseCase,
	inativar *appusuario.InativarUsuarioUseCase,
	buscarLogado *appusuario.BuscarUsuarioLogadoUseCase,
) *Handler {
	return &Handler{criar: criar, atualizar: atualizar, alterarSenha: alterarSenha, ativar: ativar, inativar: inativar, buscarLogado: buscarLogado}
}

// @Summary Cria usuário interno
// @Description Cria um usuário interno com senha inicial provisória (troca forçada no primeiro acesso). Restrito a admin.
// @Tags Usuarios
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CriarUsuarioRequest true "Dados do usuário"
// @Success 201 {object} UsuarioResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Router /v1/usuarios [post]
func (h *Handler) Criar(c *gin.Context) {
	var req CriarUsuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.criar.Executar(c.Request.Context(), toCriarInput(req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusCreated, toUsuarioResponse(out))
}

// @Summary Atualiza usuário interno
// @Description Atualiza nome, email e papel de um usuário interno. senha_nova é opcional: se informada, o admin redefine a senha (força troca no próximo login). Restrito a admin.
// @Tags Usuarios
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do usuário"
// @Param request body AtualizarUsuarioRequest true "Dados atualizados"
// @Success 200 {object} UsuarioResponse
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Router /v1/usuarios/{id} [put]
func (h *Handler) Atualizar(c *gin.Context) {
	id, ok := httprequest.ParseUintParam(c, "id")
	if !ok {
		return
	}

	var req AtualizarUsuarioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		httperror.RespondValidationError(c, "Request body inválido.")
		return
	}

	out, err := h.atualizar.Executar(c.Request.Context(), toAtualizarInput(id, req))
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.JSON(http.StatusOK, toUsuarioResponse(out))
}

// @Summary Ativa usuário interno
// @Description Reabilita um usuário interno pra login. Restrito a admin.
// @Tags Usuarios
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do usuário"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/usuarios/{id}/ativar [patch]
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

// @Summary Inativa usuário interno
// @Description Bloqueia um usuário interno de logar. Restrito a admin.
// @Tags Usuarios
// @Produce json
// @Security BearerAuth
// @Param id path int true "ID do usuário"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Router /v1/usuarios/{id}/inativar [patch]
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

// @Summary Troca a própria senha
// @Description Troca a senha do usuário logado (self-service). Usado também pra destravar o primeiro acesso (senha provisória).
// @Tags Usuarios
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body AlterarSenhaRequest true "Nova senha"
// @Success 204 "Sem conteúdo"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Router /v1/usuarios/me/senha [put]
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

	if err := h.alterarSenha.Executar(c.Request.Context(), toAlterarSenhaInput(id, req)); err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Dados do usuário logado
// @Description Retorna id/nome/email/papel/ativo do próprio usuário autenticado.
// @Tags Usuarios
// @Produce json
// @Security BearerAuth
// @Success 200 {object} UsuarioResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Router /v1/usuarios/me [get]
func (h *Handler) Me(c *gin.Context) {
	id, ok := middleware.SubjectID(c)
	if !ok {
		return
	}

	out, err := h.buscarLogado.Executar(c.Request.Context(), id)
	if err != nil {
		httperror.RespondError(c, err)
		return
	}
	c.JSON(http.StatusOK, toUsuarioResponse(out))
}
