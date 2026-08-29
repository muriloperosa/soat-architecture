//go:build integration

package integration_test

import (
	"net/http"
	"strconv"
	"testing"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
)

func abrirOrdemServicoParaEntrega(t *testing.T, adminID uint64, token string) httpordemservico.OrdemServicoResponse {
	t.Helper()

	cliente, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725", Tipo: domaincliente.TipoPessoaFisica, Nome: "Maria Silva",
		Email: "maria@email.com", Telefone: "11999998888", Senha: "senhaCliente123", CriadoPor: adminID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}
	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 52_000,
		Ano: 2020, Cor: "Prata", CriadoPor: adminID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	var aberta httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", token,
		httpordemservico.AbrirOrdemServicoRequest{ClienteID: cliente.ID, VeiculoID: veiculo.ID}, &aberta)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS: status %d, body %q", rec.Code, rec.Body.String())
	}

	return aberta
}

func TestEntregarOrdemServico_TransicaoDeFinalizadaParaEntregue(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	mecanico := seedUsuario(t, "Mecânico Oficina", "mecanico@oficina.com", "senha123", shared.PapelMecanico)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	loginMecanico := doLogin(t, "mecanico@oficina.com", "senha123")

	aberta := abrirOrdemServicoParaEntrega(t, admin.ID, loginAdmin.AccessToken)

	if err := testDB.Exec("UPDATE ordens_servico SET status = ? WHERE id = ?", "FINALIZADA", aberta.ID).Error; err != nil {
		t.Fatalf("erro ao forçar status FINALIZADA: %v", err)
	}

	var entregue httpordemservico.OrdemServicoResponse
	rec := doRequest(
		t,
		http.MethodPatch,
		"/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/entregar",
		loginMecanico.AccessToken,
		nil,
		&entregue,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("entregar OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if entregue.Status != "ENTREGUE" {
		t.Fatalf("status inesperado: %s", entregue.Status)
	}

	var quantidadeHistoricos int64
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ?", aberta.ID).
		Count(&quantidadeHistoricos).Error; err != nil {
		t.Fatalf("erro ao consultar históricos: %v", err)
	}
	if quantidadeHistoricos != 2 {
		t.Fatalf("esperados dois históricos (recebida + entregue), encontrado %d", quantidadeHistoricos)
	}

	var historico struct {
		Status      string
		AlteradoPor uint64
	}
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ? AND status = ?", aberta.ID, "ENTREGUE").
		Take(&historico).Error; err != nil {
		t.Fatalf("erro ao consultar histórico de entrega: %v", err)
	}
	if historico.AlteradoPor != mecanico.ID {
		t.Fatalf("histórico inesperado: %+v", historico)
	}

	var statusPersistido string
	if err := testDB.Table("ordens_servico").
		Select("status").
		Where("id = ?", aberta.ID).
		Scan(&statusPersistido).Error; err != nil {
		t.Fatalf("erro ao consultar status persistido: %v", err)
	}
	if statusPersistido != "ENTREGUE" {
		t.Fatalf("status não persistido: %q", statusPersistido)
	}
}

func TestEntregarOrdemServico_ValidaTransicaoDeStatus(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	aberta := abrirOrdemServicoParaEntrega(t, admin.ID, login.AccessToken)

	rec := doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/entregar",
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("entregar OS ainda RECEBIDA deveria retornar 400, veio %d", rec.Code)
	}

	if err := testDB.Exec("UPDATE ordens_servico SET status = ? WHERE id = ?", "FINALIZADA", aberta.ID).Error; err != nil {
		t.Fatalf("erro ao forçar status FINALIZADA: %v", err)
	}

	var entregue httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/entregar",
		login.AccessToken, nil, &entregue)
	if rec.Code != http.StatusOK {
		t.Fatalf("entregar OS FINALIZADA: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/entregar",
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("entregar OS já ENTREGUE deveria retornar 400, veio %d", rec.Code)
	}
}
