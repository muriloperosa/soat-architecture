package orcamento

import (
	"html/template"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

type AprovacaoEmailHandler struct {
	tokens        shared.ApprovalTokenProvider
	publicBaseURL string
	aprovar       *app.AprovarOrcamentoUseCase
	rejeitar      *app.RejeitarOrcamentoUseCase
}

func NewAprovacaoEmailHandler(
	tokens shared.ApprovalTokenProvider,
	publicBaseURL string,
	aprovar *app.AprovarOrcamentoUseCase,
	rejeitar *app.RejeitarOrcamentoUseCase,
) *AprovacaoEmailHandler {
	return &AprovacaoEmailHandler{
		tokens:        tokens,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
		aprovar:       aprovar,
		rejeitar:      rejeitar,
	}
}

var approvalConfirmationPage = template.Must(template.New("approval-confirmation").Parse(`<!doctype html>
<html lang="pt-BR">
<head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Confirmar decisão do orçamento</title></head>
<body>
  <main>
    <h1>{{if eq .Decision "APROVAR"}}Aceitar orçamento{{else}}Recusar orçamento{{end}}</h1>
    <p>Confirme sua decisão para registrar a resposta ao orçamento.</p>
    <form action="{{.PostURL}}" method="post">
      <input type="hidden" name="token" value="{{.Token}}">
      <input type="hidden" name="decisao" value="{{.Decision}}">
      {{if eq .Decision "REJEITAR"}}
      <label for="motivo">Motivo da recusa</label>
      <textarea id="motivo" name="motivo" required></textarea>
      {{end}}
      <button type="submit">Confirmar</button>
    </form>
  </main>
</body>
</html>`))

// Confirmar valida o link e apresenta uma confirmação. A abertura do GET não
// chama os casos de uso nem altera a Ordem de Serviço.
// @Summary Confirma a intenção de aprovar ou recusar um orçamento
// @Description Valida o link e apresenta um formulário de confirmação. O GET não altera a Ordem de Serviço.
// @Tags Orçamentos
// @Produce html
// @Param token query string true "Token temporário recebido por email"
// @Param acao query string true "Decisão pretendida" Enums(APROVAR, REJEITAR)
// @Success 200 {string} string "Página HTML de confirmação"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Router /v1/orcamentos/aprovacao [get]
func (h *AprovacaoEmailHandler) Confirmar(c *gin.Context) {
	acao := c.Query("acao")
	if acao != "APROVAR" && acao != "REJEITAR" {
		httperror.RespondValidationError(c, "Ação deve ser APROVAR ou REJEITAR.")
		return
	}
	if _, err := h.tokens.ValidarToken(c.Query("token")); err != nil {
		httperror.RespondUnauthorizedError(c, "Link inválido ou expirado.")
		return
	}

	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	c.Header("Content-Type", "text/html; charset=utf-8")
	if err := approvalConfirmationPage.Execute(c.Writer, struct {
		Decision string
		Token    string
		PostURL  string
	}{
		Decision: acao,
		Token:    c.Query("token"),
		PostURL:  h.publicBaseURL + "/v1/orcamentos/aprovacao/decisao",
	}); err != nil {
		return
	}
}

// Decidir valida o token e delega a mutação aos casos de uso já usados pelas
// rotas autenticadas de aprovação e rejeição.
// @Summary Registra a aprovação ou recusa de orçamento pelo link de email
// @Description Valida o token e delega a decisão aos mesmos casos de uso das rotas autenticadas. Para REJEITAR, o motivo é obrigatório.
// @Tags Orçamentos
// @Accept mpfd
// @Produce html
// @Param token formData string true "Token temporário recebido por email"
// @Param decisao formData string true "Decisão" Enums(APROVAR, REJEITAR)
// @Param motivo formData string false "Motivo obrigatório quando decisao=REJEITAR"
// @Success 200 {string} string "Página HTML de resultado"
// @Failure 400 {object} httperror.ErrorResponse
// @Failure 401 {object} httperror.ErrorResponse
// @Failure 403 {object} httperror.ErrorResponse
// @Failure 404 {object} httperror.ErrorResponse
// @Failure 409 {object} httperror.ErrorResponse
// @Failure 500 {object} httperror.ErrorResponse
// @Router /v1/orcamentos/aprovacao/decisao [post]
func (h *AprovacaoEmailHandler) Decidir(c *gin.Context) {
	var request DecisaoPorEmailRequest
	if err := c.ShouldBind(&request); err != nil {
		httperror.RespondValidationError(c, "Dados da decisão inválidos.")
		return
	}

	claims, err := h.tokens.ValidarToken(request.Token)
	if err != nil {
		httperror.RespondUnauthorizedError(c, "Link inválido ou expirado.")
		return
	}

	var output app.FluxoOrcamentoOutput
	switch request.Decisao {
	case "APROVAR":
		output, err = h.aprovar.Executar(c.Request.Context(), app.AprovarOrcamentoInput{
			OrdemServicoID: claims.OrdemServicoID,
			ClienteID:      claims.ClienteID,
		})
	case "REJEITAR":
		output, err = h.rejeitar.Executar(c.Request.Context(), app.RejeitarOrcamentoInput{
			OrdemServicoID: claims.OrdemServicoID,
			ClienteID:      claims.ClienteID,
			Motivo:         request.Motivo,
		})
	default:
		httperror.RespondValidationError(c, "Decisão deve ser APROVAR ou REJEITAR.")
		return
	}
	if err != nil {
		httperror.RespondError(c, err)
		return
	}

	c.Header("Cache-Control", "no-store")
	if strings.Contains(c.GetHeader("Accept"), "application/json") {
		c.JSON(http.StatusOK, toFluxoResponse(output))
		return
	}

	resultado := "Orçamento aprovado com sucesso."
	if request.Decisao == "REJEITAR" {
		resultado = "Orçamento recusado com sucesso."
	}
	c.Data(http.StatusOK, "text/html; charset=utf-8", []byte("<!doctype html><html lang=\"pt-BR\"><meta charset=\"utf-8\"><title>Decisão registrada</title><body><h1>"+resultado+"</h1></body></html>"))
}
