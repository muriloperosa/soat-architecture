//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
)

func TestConsultaOrdemServico_ListarBuscarPorIDEPorNumero(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	cliente, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725", Tipo: string(domaincliente.TipoPessoaFisica), Nome: "Maria Silva",
		Email: "maria.consulta@email.com", Telefone: "11999998888", Senha: "senhaCliente123", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}

	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 52000,
		Ano: 2020, Cor: "Prata", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	var aberta httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", login.AccessToken, httpordemservico.AbrirOrdemServicoRequest{
		ClienteID: cliente.ID, VeiculoID: veiculo.ID, QuilometragemEntrada: 52300, Observacoes: "Ruído no motor",
	}, &aberta)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS: status %d, body %q", rec.Code, rec.Body.String())
	}

	var listagem httpordemservico.ListarOrdensServicoResponse
	path := fmt.Sprintf("/v1/ordens-servico?page=1&status=RECEBIDA&cliente_id=%d&veiculo_id=%d", cliente.ID, veiculo.ID)
	rec = doRequest(t, http.MethodGet, path, login.AccessToken, nil, &listagem)
	if rec.Code != http.StatusOK {
		t.Fatalf("listar OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if listagem.Total != 1 || len(listagem.Items) != 1 {
		t.Fatalf("listagem inesperada: %+v", listagem)
	}
	if listagem.Page != 1 || listagem.PageSize != 20 || listagem.TotalPages != 1 {
		t.Fatalf("paginação inesperada: %+v", listagem)
	}
	if listagem.Items[0].ID != aberta.ID || listagem.Items[0].Status != "RECEBIDA" {
		t.Fatalf("OS inesperada na listagem: %+v", listagem.Items[0])
	}

	var porID httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/ordens-servico/%d", aberta.ID), login.AccessToken, nil, &porID)
	if rec.Code != http.StatusOK {
		t.Fatalf("buscar OS por ID: status %d, body %q", rec.Code, rec.Body.String())
	}
	if porID.ID != aberta.ID || len(porID.HistoricoStatus) != 1 {
		t.Fatalf("consulta por ID inesperada: %+v", porID)
	}

	var porNumero httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico/numero/"+aberta.Numero, login.AccessToken, nil, &porNumero)
	if rec.Code != http.StatusOK {
		t.Fatalf("buscar OS por número: status %d, body %q", rec.Code, rec.Body.String())
	}
	if porNumero.ID != aberta.ID || porNumero.Numero != aberta.Numero {
		t.Fatalf("consulta por número inesperada: %+v", porNumero)
	}
}
