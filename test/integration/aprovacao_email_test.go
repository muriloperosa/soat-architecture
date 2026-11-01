//go:build integration

package integration_test

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
	"github.com/stretchr/testify/require"
)

func TestAprovacaoEmail_EmittedLinkConfirmsApprovalEndToEnd(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")
	os, links := prepararOrcamentoComLinksEmail(t, admin.ID, login.AccessToken)
	approveURL := links["APROVAR"]
	path := approveURL.RequestURI()

	getResponse := doRequest(t, http.MethodGet, path, "", nil, nil)
	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	require.Contains(t, getResponse.Body.String(), "name=\"decisao\" value=\"APROVAR\"")
	require.Equal(t, "no-store", getResponse.Header().Get("Cache-Control"))
	require.Equal(t, "AGUARDANDO_APROVACAO", statusOSPersistido(t, os.ID), "abrir o link via GET nao deve alterar o estado")

	form := url.Values{"token": {approveURL.Query().Get("token")}, "decisao": {"APROVAR"}}
	postResponse := postDecisionIntegration(t, form)
	require.Equal(t, http.StatusOK, postResponse.Code, postResponse.Body.String())
	require.Contains(t, postResponse.Body.String(), "Orçamento aprovado")
	require.Equal(t, "APROVADA", statusOSPersistido(t, os.ID))

	var historicoCount int64
	require.NoError(t, testDB.Table("historicos_status").Where("ordem_servico_id = ? AND status = ?", os.ID, "APROVADA").Count(&historicoCount).Error)
	require.EqualValues(t, 1, historicoCount, "a decisao pelo link deve ser persistida no historico")
}

func TestAprovacaoEmail_EmittedLinkRejectsWithRequiredReasonEndToEnd(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")
	os, links := prepararOrcamentoComLinksEmail(t, admin.ID, login.AccessToken)
	rejectURL := links["REJEITAR"]

	getResponse := doRequest(t, http.MethodGet, rejectURL.RequestURI(), "", nil, nil)
	require.Equal(t, http.StatusOK, getResponse.Code, getResponse.Body.String())
	require.Contains(t, getResponse.Body.String(), "name=\"motivo\"")
	require.Equal(t, "AGUARDANDO_APROVACAO", statusOSPersistido(t, os.ID))

	baseForm := url.Values{"token": {rejectURL.Query().Get("token")}, "decisao": {"REJEITAR"}}
	withoutReason := postDecisionIntegration(t, baseForm)
	require.Equal(t, http.StatusBadRequest, withoutReason.Code, withoutReason.Body.String())
	require.Equal(t, "AGUARDANDO_APROVACAO", statusOSPersistido(t, os.ID), "rejeicao sem motivo nao deve alterar a OS")

	baseForm.Set("motivo", "Cliente solicita revisao do valor")
	postResponse := postDecisionIntegration(t, baseForm)
	require.Equal(t, http.StatusOK, postResponse.Code, postResponse.Body.String())
	require.Contains(t, postResponse.Body.String(), "Orçamento recusado")
	require.Equal(t, "REJEITADA", statusOSPersistido(t, os.ID))

	var motivo string
	require.NoError(t, testDB.Table("historicos_status").Select("motivo").Where("ordem_servico_id = ? AND status = ?", os.ID, "REJEITADA").Scan(&motivo).Error)
	require.Equal(t, "Cliente solicita revisao do valor", motivo)
}

func prepararOrcamentoComLinksEmail(t *testing.T, adminID uint64, token string) (httpordemservico.OrdemServicoResponse, map[string]*url.URL) {
	t.Helper()

	os := abrirOrdemServicoParaExecucao(t, adminID, token)
	osPath := "/v1/ordens-servico/" + strconv.FormatUint(os.ID, 10)
	require.Equal(t, http.StatusOK, doRequest(t, http.MethodPatch, osPath+"/iniciar-diagnostico", token, nil, nil).Code)
	require.Equal(t, http.StatusOK, doRequest(t, http.MethodPut, osPath+"/diagnostico", token,
		map[string]string{"diagnostico": "Troca de oleo necessaria"}, nil).Code)

	service, err := testContainer.CriarServicoUC.Executar(t.Context(), appservico.CriarServicoInput{
		Nome: "Troca de oleo", Descricao: "Troca de oleo do motor", PrecoBase: 120,
		TempoEstimadoMinutos: 45, CriadoPor: adminID,
	})
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, doRequest(t, http.MethodPost, osPath+"/orcamento", token,
		httporcamento.GerarOrcamentoRequest{Observacoes: "Teste integrado de links"}, nil).Code)
	require.Equal(t, http.StatusOK, doRequest(t, http.MethodPost, osPath+"/orcamento/itens-servico", token,
		httporcamento.AdicionarServicoOrcamentoRequest{ServicoID: service.ID, Quantidade: 1}, nil).Code)

	var emailLog bytes.Buffer
	previousLogWriter := log.Writer()
	log.SetOutput(&emailLog)
	defer log.SetOutput(previousLogWriter)
	sendResponse := doRequest(t, http.MethodPatch, osPath+"/orcamento/enviar-aprovacao", token, nil, nil)
	require.Equal(t, http.StatusOK, sendResponse.Code, sendResponse.Body.String())

	links := make(map[string]*url.URL, 2)
	for _, line := range strings.Split(emailLog.String(), "\n") {
		for action, marker := range map[string]string{"APROVAR": "Aceitar orçamento: ", "REJEITAR": "Recusar orçamento: "} {
			index := strings.Index(line, marker)
			if index < 0 {
				continue
			}
			link, err := url.Parse(strings.TrimSpace(line[index+len(marker):]))
			require.NoError(t, err)
			require.Equal(t, action, link.Query().Get("acao"))
			require.NotEmpty(t, link.Query().Get("token"))
			links[action] = link
		}
	}
	require.Contains(t, links, "APROVAR", "email simulado deve conter link de aprovacao")
	require.Contains(t, links, "REJEITAR", "email simulado deve conter link de rejeicao")
	require.Equal(t, links["APROVAR"].Query().Get("token"), links["REJEITAR"].Query().Get("token"))
	return os, links
}

func postDecisionIntegration(t *testing.T, values url.Values) *httptest.ResponseRecorder {
	t.Helper()
	request := httptest.NewRequest(http.MethodPost, "/v1/orcamentos/aprovacao/decisao", strings.NewReader(values.Encode()))
	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	response := httptest.NewRecorder()
	testRouter.ServeHTTP(response, request)
	return response
}

func statusOSPersistido(t *testing.T, id uint64) string {
	t.Helper()
	var status string
	require.NoError(t, testDB.Table("ordens_servico").Select("status").Where("id = ?", id).Scan(&status).Error)
	return status
}
