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
		{"REJEITADA", base.Add(time.Hour)}, // Mesmo nível de APROVADA, desempata pela data.
		{"EM_EXECUCAO", base},
		{"FINALIZADA", base},
		{"ENTREGUE", base},
	}

	ids := make([]uint64, 0, len(tipoOrdem))
	baseID := seedOrdemServico(t, admin.ID)
	for i, item := range tipoOrdem {
		id := baseID
		if i > 0 {
			numero := fmt.Sprintf("OS-LIST-%d-%d", time.Now().UnixNano(), i)
			if err := testDB.Exec(`
				INSERT INTO ordens_servico
					(numero, cliente_id, veiculo_id, quilometragem_entrada, status, criado_por, data_cadastro, data_atualizacao)
				SELECT ?, cliente_id, veiculo_id, quilometragem_entrada, 'RECEBIDA', criado_por, ?, ?
				FROM ordens_servico WHERE id = ?`, numero, item.data, item.data, baseID).Error; err != nil {
				t.Fatalf("duplicar OS para cenário de listagem: %v", err)
			}
			if err := testDB.Raw("SELECT id FROM ordens_servico WHERE numero = ?", numero).Scan(&id).Error; err != nil {
				t.Fatalf("buscar OS duplicada: %v", err)
			}
		}
		if err := testDB.Exec("UPDATE ordens_servico SET status = ?, data_cadastro = ? WHERE id = ?", item.status, item.data, id).Error; err != nil {
			t.Fatalf("preparar OS %d com status %s: %v", id, item.status, err)
		}
		ids = append(ids, id)
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
		{ids[1], "RECEBIDA"},
		{ids[0], "RECEBIDA"},
		{ids[2], "EM_DIAGNOSTICO"},
		{ids[3], "AGUARDANDO_APROVACAO"},
		{ids[5], "REJEITADA"},
		{ids[4], "APROVADA"},
		{ids[6], "EM_EXECUCAO"},
	}
	for i, esperado := range esperados {
		require.Equal(t, esperado.id, listagem.Items[i].ID, fmt.Sprintf("posição %d", i))
		require.Equal(t, esperado.status, listagem.Items[i].Status, fmt.Sprintf("posição %d", i))
	}
}
