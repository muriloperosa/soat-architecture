//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	httpservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/servico"
	"github.com/stretchr/testify/require"
)

// TestOrdemServicoFluxoCompleto_DaAberturaAEntregaPassandoPorTodasAsEtapas
// percorre, exclusivamente via HTTP, o ciclo de vida inteiro de uma Ordem de
// Serviço: abertura -> diagnóstico -> orçamento (serviço + peça) -> envio
// para aprovação -> aprovação do cliente (que cria a reserva automática) ->
// início de execução -> finalização (baixa o estoque) -> entrega. Confirma
// o estado ao final de cada etapa, não só o resultado final.
func TestOrdemServicoFluxoCompleto_DaAberturaAEntregaPassandoPorTodasAsEtapas(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	mecanico := seedUsuario(t, "Mecânico Oficina", "mecanico@oficina.com", "senha123", shared.PapelMecanico)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	loginMecanico := doLogin(t, "mecanico@oficina.com", "senha123")

	// ---------------------------------------------------------------------
	// Abertura: RECEBIDA
	// ---------------------------------------------------------------------
	aberta := abrirOrdemServicoParaExecucao(t, admin.ID, loginAdmin.AccessToken)
	if aberta.Status != "RECEBIDA" {
		t.Fatalf("abertura: status inesperado %q", aberta.Status)
	}
	osPath := "/v1/ordens-servico/" + strconv.FormatUint(aberta.ID, 10)

	loginCliente := doLoginCliente(t, "maria@email.com", "senhaCliente123")

	// Catálogo: um serviço e uma peça com estoque suficiente pro orçamento.
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 2)

	var servicoResp struct {
		ID uint64 `json:"id"`
	}
	rec := doRequest(t, http.MethodPost, "/v1/servicos", loginAdmin.AccessToken,
		httpservico.CriarServicoRequest{Nome: "Troca de óleo", Descricao: "Troca de óleo do motor", PrecoBase: floatPtr(100.0), TempoEstimadoMinutos: 60},
		&servicoResp)
	if rec.Code != http.StatusCreated {
		t.Fatalf("criar serviço: status %d, body %q", rec.Code, rec.Body.String())
	}

	// ---------------------------------------------------------------------
	// Diagnóstico: RECEBIDA -> EM_DIAGNOSTICO
	// ---------------------------------------------------------------------
	rec = doRequest(t, http.MethodPatch, osPath+"/iniciar-diagnostico", loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, http.MethodPut, osPath+"/diagnostico", loginAdmin.AccessToken,
		httpordemservico.InformarDiagnosticoRequest{Diagnostico: "Necessária troca de óleo e pastilha de freio"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("informar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}

	// ---------------------------------------------------------------------
	// Orçamento: gera, adiciona item de serviço e item de peça
	// ---------------------------------------------------------------------
	var orcamento httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento", loginAdmin.AccessToken,
		httporcamento.GerarOrcamentoRequest{Observacoes: "Orçamento do fluxo completo"}, &orcamento)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gerar orçamento: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPost, osPath+"/orcamento/itens-servico", loginAdmin.AccessToken,
		httporcamento.AdicionarServicoOrcamentoRequest{ServicoID: servicoResp.ID, Quantidade: 1}, &orcamento)
	if rec.Code != http.StatusOK {
		t.Fatalf("adicionar item de serviço: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPost, osPath+"/orcamento/itens-peca", loginAdmin.AccessToken,
		httporcamento.AdicionarPecaOrcamentoRequest{PecaID: pecaID, Quantidade: 2}, &orcamento)
	if rec.Code != http.StatusOK {
		t.Fatalf("adicionar item de peça: status %d, body %q", rec.Code, rec.Body.String())
	}
	if orcamento.ValorTotal != 279.8 {
		t.Fatalf("valor total do orçamento inesperado: %.2f", orcamento.ValorTotal)
	}

	// ---------------------------------------------------------------------
	// Envio para aprovação: EM_DIAGNOSTICO -> AGUARDANDO_APROVACAO
	// ---------------------------------------------------------------------
	rec = doRequest(t, http.MethodPatch, osPath+"/orcamento/enviar-aprovacao", loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("enviar para aprovação: status %d, body %q", rec.Code, rec.Body.String())
	}

	// ---------------------------------------------------------------------
	// Aprovação do cliente: AGUARDANDO_APROVACAO -> APROVADA
	// Aqui nasce a ReservaPeca automaticamente.
	// ---------------------------------------------------------------------
	var aprovada httporcamento.FluxoOrcamentoResponse
	rec = doRequest(t, http.MethodPatch, osPath+"/orcamento/aprovar", loginCliente.AccessToken, nil, &aprovada)
	if rec.Code != http.StatusOK {
		t.Fatalf("aprovar orçamento: status %d, body %q", rec.Code, rec.Body.String())
	}
	if aprovada.Status != "APROVADA" {
		t.Fatalf("status pós-aprovação inesperado: %q", aprovada.Status)
	}

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 2, reservada, "aprovação deveria reservar a quantidade do orçamento")

	// ---------------------------------------------------------------------
	// Início de execução: APROVADA -> EM_EXECUCAO
	// ---------------------------------------------------------------------
	var emExecucao httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodPatch, osPath+"/iniciar-execucao", loginMecanico.AccessToken, nil, &emExecucao)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar execução: status %d, body %q", rec.Code, rec.Body.String())
	}
	if emExecucao.Status != "EM_EXECUCAO" {
		t.Fatalf("status pós-início de execução inesperado: %q", emExecucao.Status)
	}

	// ---------------------------------------------------------------------
	// Finalização: EM_EXECUCAO -> FINALIZADA
	// Consome o estoque físico e libera a reserva.
	// ---------------------------------------------------------------------
	var finalizada httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodPatch, osPath+"/finalizar", loginMecanico.AccessToken, nil, &finalizada)
	if rec.Code != http.StatusOK {
		t.Fatalf("finalizar OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if finalizada.Status != "FINALIZADA" {
		t.Fatalf("status pós-finalização inesperado: %q", finalizada.Status)
	}

	peca, err := testContainer.PecaRepo.BuscarPorID(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 8, peca.QuantidadeEmEstoque(), "finalização deveria baixar o estoque físico em 2")

	reservada, err = testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Zero(t, reservada, "finalização deveria zerar a reserva")

	// ---------------------------------------------------------------------
	// Entrega: FINALIZADA -> ENTREGUE
	// ---------------------------------------------------------------------
	var entregue httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodPatch, osPath+"/entregar", loginMecanico.AccessToken, nil, &entregue)
	if rec.Code != http.StatusOK {
		t.Fatalf("entregar OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if entregue.Status != "ENTREGUE" {
		t.Fatalf("status final inesperado: %q", entregue.Status)
	}

	// ---------------------------------------------------------------------
	// Histórico completo: uma entrada por transição de status
	// RECEBIDA, EM_DIAGNOSTICO, AGUARDANDO_APROVACAO, APROVADA,
	// EM_EXECUCAO, FINALIZADA, ENTREGUE
	// ---------------------------------------------------------------------
	var statusHistoricos []string
	if err := testDB.Table("historicos_status").
		Select("status").
		Where("ordem_servico_id = ?", aberta.ID).
		Order("id ASC").
		Pluck("status", &statusHistoricos).Error; err != nil {
		t.Fatalf("erro ao consultar históricos: %v", err)
	}
	esperado := []string{"RECEBIDA", "EM_DIAGNOSTICO", "AGUARDANDO_APROVACAO", "APROVADA", "EM_EXECUCAO", "FINALIZADA", "ENTREGUE"}
	require.Equal(t, esperado, statusHistoricos, "sequência de histórico de status inesperada")

	// Quem alterou o status pra EM_EXECUCAO/FINALIZADA/ENTREGUE precisa ser o mecânico.
	var alteradoPorFinalizacao uint64
	if err := testDB.Table("historicos_status").
		Select("alterado_por").
		Where("ordem_servico_id = ? AND status = ?", aberta.ID, "FINALIZADA").
		Scan(&alteradoPorFinalizacao).Error; err != nil {
		t.Fatalf("erro ao consultar histórico de finalização: %v", err)
	}
	require.Equal(t, mecanico.ID, alteradoPorFinalizacao)
}
