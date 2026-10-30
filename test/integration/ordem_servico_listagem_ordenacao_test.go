//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
	"github.com/stretchr/testify/require"
)

func TestConsultaOrdemServico_ListagemOrdenaPorStatusEExcluiFinalizadas(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	base := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tipoOrdem := []struct {
		status string
		data   time.Time
	}{
		{"RECEBIDA", base.Add(2 * time.Hour)},
		{"RECEBIDA", base.Add(time.Hour)}, // A mais antiga deve aparecer primeiro.
		{"EM_DIAGNOSTICO", base},
		{"AGUARDANDO_APROVACAO", base},
		{"APROVADA", base.Add(2 * time.Hour)},
		{"REJEITADA", base.Add(time.Hour)}, // Mesmo nível de AGUARDANDO_APROVACAO, desempata pela data.
		{"EM_EXECUCAO", base},
		{"ENTREGUE", base},
	}

	baseID := seedOrdemServico(t, admin.ID)
	if err := testDB.Exec("UPDATE ordens_servico SET status = ?, data_cadastro = ? WHERE id = ?", "FINALIZADA", base, baseID).Error; err != nil {
		t.Fatalf("preparar OS finalizada: %v", err)
	}

	idsPorStatus := make(map[string][]uint64)
	for _, item := range tipoOrdem {
		id := seedOrdemServicoComStatus(t, baseID, item.status, item.data)
		idsPorStatus[item.status] = append(idsPorStatus[item.status], id)
	}

	var listagem httpordemservico.ListarOrdensServicoResponse
	rec := doRequest(t, http.MethodGet, "/v1/ordens-servico", login.AccessToken, nil, &listagem)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, int64(7), listagem.Total)
	require.Len(t, listagem.Items, 7)

	esperados := []struct {
		id     uint64
		status string
	}{
		{idsPorStatus["EM_EXECUCAO"][0], "EM_EXECUCAO"},
		{idsPorStatus["APROVADA"][0], "APROVADA"},
		{idsPorStatus["AGUARDANDO_APROVACAO"][0], "AGUARDANDO_APROVACAO"},
		{idsPorStatus["REJEITADA"][0], "REJEITADA"},
		{idsPorStatus["EM_DIAGNOSTICO"][0], "EM_DIAGNOSTICO"},
		{idsPorStatus["RECEBIDA"][1], "RECEBIDA"},
		{idsPorStatus["RECEBIDA"][0], "RECEBIDA"},
	}
	for i, esperado := range esperados {
		require.Equal(t, esperado.id, listagem.Items[i].ID, fmt.Sprintf("posição %d", i))
		require.Equal(t, esperado.status, listagem.Items[i].Status, fmt.Sprintf("posição %d", i))
	}

	var listagemPorStatus httpordemservico.ListarOrdensServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico?order=status", login.AccessToken, nil, &listagemPorStatus)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	for i, esperado := range esperados {
		require.Equal(t, esperado.id, listagemPorStatus.Items[i].ID, fmt.Sprintf("order=status, posição %d", i))
	}

	var listagemConcluidas httpordemservico.ListarOrdensServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico?status=FINALIZADA,ENTREGUE", login.AccessToken, nil, &listagemConcluidas)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, int64(2), listagemConcluidas.Total)
	statusConcluido := map[string]bool{}
	for _, item := range listagemConcluidas.Items {
		statusConcluido[item.Status] = true
	}
	require.True(t, statusConcluido["FINALIZADA"])
	require.True(t, statusConcluido["ENTREGUE"])

	var emailCliente string
	if err := testDB.Raw(`
		SELECT clientes.email
		FROM clientes
		INNER JOIN ordens_servico ON ordens_servico.cliente_id = clientes.id
		WHERE ordens_servico.id = ?`, baseID).Scan(&emailCliente).Error; err != nil {
		t.Fatalf("buscar email do cliente dono das OS: %v", err)
	}
	loginCliente := doLoginCliente(t, emailCliente, "senha123")
	var historicoCliente httpordemservico.ListarOrdensServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico?status=FINALIZADA,ENTREGUE", loginCliente.AccessToken, nil, &historicoCliente)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	require.Equal(t, int64(2), historicoCliente.Total)
	require.Len(t, historicoCliente.Items, 2)

	var listagemDescendente httpordemservico.ListarOrdensServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico?direction=DESC", login.AccessToken, nil, &listagemDescendente)
	require.Equal(t, http.StatusOK, rec.Code, rec.Body.String())
	idsDescendentes := []uint64{
		idsPorStatus["RECEBIDA"][0],
		idsPorStatus["RECEBIDA"][1],
		idsPorStatus["EM_DIAGNOSTICO"][0],
		idsPorStatus["REJEITADA"][0],
		idsPorStatus["AGUARDANDO_APROVACAO"][0],
		idsPorStatus["APROVADA"][0],
		idsPorStatus["EM_EXECUCAO"][0],
	}
	for i, esperadoID := range idsDescendentes {
		require.Equal(t, esperadoID, listagemDescendente.Items[i].ID, fmt.Sprintf("direction=DESC, posição %d", i))
	}
}

func seedOrdemServicoComStatus(t *testing.T, ordemTemplateID uint64, status string, dataCadastro time.Time) uint64 {
	t.Helper()

	numero := fmt.Sprintf("OS-LIST-%d", time.Now().UnixNano())
	if err := testDB.Exec(`
		INSERT INTO ordens_servico
			(numero, cliente_id, veiculo_id, quilometragem_entrada, status, criado_por, data_cadastro, data_atualizacao)
		SELECT ?, cliente_id, veiculo_id, quilometragem_entrada, ?, criado_por, ?, ?
		FROM ordens_servico WHERE id = ?`, numero, status, dataCadastro, dataCadastro, ordemTemplateID).Error; err != nil {
		t.Fatalf("semear OS com status %s: %v", status, err)
	}

	var id uint64
	if err := testDB.Raw("SELECT id FROM ordens_servico WHERE numero = ?", numero).Scan(&id).Error; err != nil {
		t.Fatalf("buscar OS semeada com status %s: %v", status, err)
	}
	return id
}
