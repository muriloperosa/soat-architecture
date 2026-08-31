//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"testing"

	apporcamento "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	appordemservico "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
	"github.com/stretchr/testify/require"
)

func forcarStatusOrdemServico(t *testing.T, id uint64, status string) {
	t.Helper()
	if err := testDB.Exec("UPDATE ordens_servico SET status = ? WHERE id = ?", status, id).Error; err != nil {
		t.Fatalf("erro ao forçar status %s: %v", status, err)
	}
}

// aprovarEIniciarExecucao leva uma OS RECEBIDA (com orçamento contendo a
// peça informada) até EM_EXECUCAO, passando pelo fluxo real de aprovação
// que é quem cria a ReservaPeca automaticamente (AprovarOrcamentoUseCase).
func aprovarEIniciarExecucao(t *testing.T, ordemServicoID, usuarioID, pecaID uint64, quantidade int) {
	t.Helper()
	ctx := context.Background()

	prepararOrcamentoComPecaAguardando(t, ordemServicoID, usuarioID, pecaID, quantidade)

	clienteID := clienteIDDaOS(t, ordemServicoID)
	_, err := testContainer.AprovarOrcamentoUC.Executar(ctx, apporcamento.AprovarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ClienteID:      clienteID,
	})
	require.NoError(t, err)

	_, err = testContainer.IniciarExecucaoUC.Executar(ctx, appordemservico.IniciarExecucaoInput{
		OrdemServicoID: ordemServicoID,
		UsuarioID:      usuarioID,
	})
	require.NoError(t, err)
}

func TestFinalizarOrdemServico_ConsomeReservasETransicionaParaFinalizada(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	mecanico := seedUsuario(t, "Mecânico Oficina", "mecanico@oficina.com", "senha123", shared.PapelMecanico)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	loginMecanico := doLogin(t, "mecanico@oficina.com", "senha123")

	aberta := abrirOrdemServicoParaExecucao(t, admin.ID, loginAdmin.AccessToken)
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 2)
	aprovarEIniciarExecucao(t, aberta.ID, admin.ID, pecaID, 3)

	var finalizada httpordemservico.OrdemServicoResponse
	rec := doRequest(
		t,
		http.MethodPatch,
		"/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/finalizar",
		loginMecanico.AccessToken,
		nil,
		&finalizada,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("finalizar OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if finalizada.Status != "FINALIZADA" {
		t.Fatalf("status inesperado: %s", finalizada.Status)
	}

	peca, err := testContainer.PecaRepo.BuscarPorID(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 7, peca.QuantidadeEmEstoque())

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Zero(t, reservada)

	var quantidadeHistoricos int64
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ? AND status = ?", aberta.ID, "FINALIZADA").
		Count(&quantidadeHistoricos).Error; err != nil {
		t.Fatalf("erro ao consultar históricos: %v", err)
	}
	if quantidadeHistoricos != 1 {
		t.Fatalf("esperado 1 histórico FINALIZADA, encontrado %d", quantidadeHistoricos)
	}

	var historico struct {
		AlteradoPor uint64
	}
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ? AND status = ?", aberta.ID, "FINALIZADA").
		Take(&historico).Error; err != nil {
		t.Fatalf("erro ao consultar histórico de finalização: %v", err)
	}
	if historico.AlteradoPor != mecanico.ID {
		t.Fatalf("histórico inesperado: %+v", historico)
	}
}

func TestFinalizarOrdemServico_SomenteEmExecucao(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")
	aberta := abrirOrdemServicoParaExecucao(t, admin.ID, login.AccessToken)

	rec := doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/finalizar",
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("finalizar OS RECEBIDA deveria retornar 400, veio %d", rec.Code)
	}
}

func TestFinalizarOrdemServico_PapelNaoAutorizadoRetorna403(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	seedUsuario(t, "Atendente Oficina", "atendente@oficina.com", "senha123", shared.PapelAtendente)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	loginAtendente := doLogin(t, "atendente@oficina.com", "senha123")

	aberta := abrirOrdemServicoParaExecucao(t, admin.ID, loginAdmin.AccessToken)
	forcarStatusOrdemServico(t, aberta.ID, "EM_EXECUCAO")

	rec := doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/finalizar",
		loginAtendente.AccessToken, nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("finalizar por atendente deveria ser 403, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestFinalizarOrdemServico_EstoqueAbaixoDoMinimoFazRollback(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	aberta := abrirOrdemServicoParaExecucao(t, admin.ID, login.AccessToken)
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 2)
	aprovarEIniciarExecucao(t, aberta.ID, admin.ID, pecaID, 5)

	if err := testDB.Exec("UPDATE pecas SET quantidade_em_estoque = ? WHERE id = ?", 4, pecaID).Error; err != nil {
		t.Fatalf("erro ao forçar estoque incompatível: %v", err)
	}

	rec := doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/finalizar",
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("consumo abaixo do mínimo deveria retornar 400, veio %d, body %q", rec.Code, rec.Body.String())
	}

	var statusPersistido string
	if err := testDB.Table("ordens_servico").Select("status").Where("id = ?", aberta.ID).Scan(&statusPersistido).Error; err != nil {
		t.Fatalf("erro ao consultar status persistido: %v", err)
	}
	if statusPersistido != "EM_EXECUCAO" {
		t.Fatalf("rollback deveria manter EM_EXECUCAO, ficou %q", statusPersistido)
	}

	peca, err := testContainer.PecaRepo.BuscarPorID(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 4, peca.QuantidadeEmEstoque(), "estoque físico não pode ter sido baixado após rollback")

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 5, reservada, "reserva deve permanecer após rollback")

	var historicoFinalizada int64
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ? AND status = ?", aberta.ID, "FINALIZADA").
		Count(&historicoFinalizada).Error; err != nil {
		t.Fatalf("erro ao consultar históricos: %v", err)
	}
	if historicoFinalizada != 0 {
		t.Fatalf("não deveria existir histórico FINALIZADA após rollback, encontrado %d", historicoFinalizada)
	}
}

func TestFinalizarOrdemServico_ConcorrenciaNaMesmaPecaSerializaConsumo(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	pecaID := seedPecaComEstoque(t, admin.ID, 10, 0)

	os1 := seedOrdemServico(t, admin.ID)
	aprovarEIniciarExecucao(t, os1, admin.ID, pecaID, 5)

	// os2 reaproveita o cliente/veículo de os1 (seedOrdemServico usa um CPF
	// fixo, então chamá-la de novo bateria em "cliente já cadastrado").
	var clienteID, veiculoID uint64
	if err := testDB.Raw("SELECT cliente_id, veiculo_id FROM ordens_servico WHERE id = ?", os1).Row().Scan(&clienteID, &veiculoID); err != nil {
		t.Fatalf("erro ao ler OS semeada: %v", err)
	}
	if err := testDB.Exec(
		"INSERT INTO ordens_servico (numero, cliente_id, veiculo_id, quilometragem_entrada, status, criado_por) VALUES (?, ?, ?, 0, 'RECEBIDA', ?)",
		"OS-CONC-0002", clienteID, veiculoID, admin.ID,
	).Error; err != nil {
		t.Fatalf("erro ao semear segunda OS: %v", err)
	}
	var os2 uint64
	if err := testDB.Raw("SELECT id FROM ordens_servico WHERE numero = ?", "OS-CONC-0002").Scan(&os2).Error; err != nil {
		t.Fatalf("erro ao buscar segunda OS: %v", err)
	}
	aprovarEIniciarExecucao(t, os2, admin.ID, pecaID, 5)

	var wg sync.WaitGroup
	erros := make(chan error, 2)
	for _, osID := range []uint64{os1, os2} {
		wg.Add(1)
		go func(id uint64) {
			defer wg.Done()
			rec := doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(id, 10)+"/finalizar",
				login.AccessToken, nil, nil)
			if rec.Code != http.StatusOK {
				erros <- &httpStatusError{code: rec.Code, body: rec.Body.String()}
			}
		}(osID)
	}
	wg.Wait()
	close(erros)

	for err := range erros {
		t.Fatalf("finalização concorrente falhou: %v", err)
	}

	peca, err := testContainer.PecaRepo.BuscarPorID(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 0, peca.QuantidadeEmEstoque(), "as duas finalizações devem baixar 5+5 sem lost update")
	require.GreaterOrEqual(t, peca.QuantidadeEmEstoque(), peca.EstoqueMinimo())

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Zero(t, reservada)
}

type httpStatusError struct {
	code int
	body string
}

func (e *httpStatusError) Error() string {
	return "status " + strconv.Itoa(e.code) + ": " + e.body
}
