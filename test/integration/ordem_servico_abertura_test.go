//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
)

func TestAbrirOrdemServico_PersisteOSHistoricoERetorna201(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	cliente, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Maria Silva",
		Email:     "maria@email.com",
		Telefone:  "11999998888",
		Senha:     "senhaCliente123",
		CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}

	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa:              "ABC1D23",
		Marca:              "Fiat",
		Modelo:             "Uno",
		QuilometragemAtual: 52_000,
		Ano:                2020,
		Cor:                "Prata",
		CriadoPor:          admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	var response httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", login.AccessToken,
		httpordemservico.AbrirOrdemServicoRequest{
			ClienteID:            cliente.ID,
			VeiculoID:            veiculo.ID,
			QuilometragemEntrada: 52_300,
			Observacoes:          "Ruído no motor",
		},
		&response,
	)

	if rec.Code != http.StatusCreated {
		t.Fatalf("Abrir OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if response.ID == 0 || response.Numero == "" || response.Status != "RECEBIDA" {
		t.Fatalf("resposta inesperada: %+v", response)
	}
	if response.ClienteID != cliente.ID || response.VeiculoID != veiculo.ID || response.QuilometragemEntrada != 52_300 {
		t.Fatalf("dados da OS não foram preservados: %+v", response)
	}
	if len(response.HistoricoStatus) != 1 || response.HistoricoStatus[0].ID == 0 {
		t.Fatalf("histórico inicial deveria vir com ID real do banco, veio: %+v", response.HistoricoStatus)
	}

	var quantidadeHistoricos int64
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ? AND status = ? AND alterado_por = ?", response.ID, "RECEBIDA", admin.ID).
		Count(&quantidadeHistoricos).Error; err != nil {
		t.Fatalf("erro ao consultar histórico: %v", err)
	}
	if quantidadeHistoricos != 1 {
		t.Fatalf("esperado um histórico RECEBIDA, encontrado %d", quantidadeHistoricos)
	}
}

func TestAbrirOrdemServico_ClienteInexistenteRetorna404(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	rec := doRequest(
		t, http.MethodPost, "/v1/ordens-servico", login.AccessToken,
		httpordemservico.AbrirOrdemServicoRequest{ClienteID: 999, VeiculoID: 999}, nil,
	)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("cliente inexistente deveria retornar 404, veio %d, body %q", rec.Code, rec.Body.String())
	}
}
