//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpveiculo "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/veiculo"
)

func TestVeiculoLifecycle_TodosOsEndpoints(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	var criado httpveiculo.VeiculoResponse
	rec := doRequest(t, http.MethodPost, "/v1/veiculos", loginAdmin.AccessToken,
		httpveiculo.CadastrarVeiculoRequest{Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 15000, Ano: 2020, Cor: "Prata"},
		&criado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Cadastrar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if criado.Placa != "ABC1D23" || !criado.Ativo || criado.CriadoPor == 0 {
		t.Fatalf("Cadastrar: resposta inesperada: %+v", criado)
	}

	var listagem httpveiculo.ListarVeiculosResponse
	rec = doRequest(t, http.MethodGet, "/v1/veiculos", loginAdmin.AccessToken, nil, &listagem)

	if rec.Code != http.StatusOK {
		t.Fatalf("Listar: status %d, body %q", rec.Code, rec.Body.String())
	}

	if listagem.Total != 1 || len(listagem.Items) != 1 || listagem.Items[0].ID != criado.ID {
		t.Fatalf("Listar: resposta inesperada: %+v", listagem)
	}

	if listagem.Offset != 0 || listagem.Limit != 20 || listagem.Order != "id" || listagem.Direction != "ASC" {
		t.Fatalf("Listar: paginação inesperada: %+v", listagem)
	}

	var consultado httpveiculo.VeiculoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/veiculos/%d", criado.ID), loginAdmin.AccessToken, nil, &consultado)
	if rec.Code != http.StatusOK || consultado.ID != criado.ID || consultado.QuilometragemAtual != 15000 {
		t.Fatalf("ConsultarPorID: status %d, body %+v", rec.Code, consultado)
	}

	var consultadoPorPlaca httpveiculo.VeiculoResponse
	rec = doRequest(t, http.MethodGet, "/v1/veiculos/placa/"+criado.Placa, loginAdmin.AccessToken, nil, &consultadoPorPlaca)
	if rec.Code != http.StatusOK || consultadoPorPlaca.ID != criado.ID {
		t.Fatalf("ConsultarPorPlaca: status %d, body %+v", rec.Code, consultadoPorPlaca)
	}

	var atualizado httpveiculo.VeiculoResponse
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/v1/veiculos/%d", criado.ID), loginAdmin.AccessToken,
		httpveiculo.AtualizarVeiculoRequest{Marca: "Volkswagen", Modelo: "Gol", Cor: "Preto", QuilometragemAtual: 16000},
		&atualizado)
	if rec.Code != http.StatusOK {
		t.Fatalf("Atualizar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if atualizado.Marca != "Volkswagen" || atualizado.Modelo != "Gol" || atualizado.Cor != "Preto" || atualizado.Placa != "ABC1D23" || atualizado.QuilometragemAtual != 16000 {
		t.Fatalf("Atualizar: dados não bateram: %+v", atualizado)
	}

	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/v1/veiculos/%d", criado.ID), loginAdmin.AccessToken,
		httpveiculo.AtualizarVeiculoRequest{Marca: "Volkswagen", Modelo: "Gol", Cor: "Preto", QuilometragemAtual: 10000},
		nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("Atualizar com quilometragem menor que a atual deveria ser 400, veio %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/veiculos/%d/inativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Inativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	var posInativar httpveiculo.VeiculoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/veiculos/%d", criado.ID), loginAdmin.AccessToken, nil, &posInativar)
	if rec.Code != http.StatusOK || posInativar.Ativo {
		t.Fatalf("ConsultarPorID pós-Inativar deveria vir ativo=false: %+v", posInativar)
	}

	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/veiculos/%d/ativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Ativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	var posAtivar httpveiculo.VeiculoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/veiculos/%d", criado.ID), loginAdmin.AccessToken, nil, &posAtivar)
	if rec.Code != http.StatusOK || !posAtivar.Ativo {
		t.Fatalf("ConsultarPorID pós-Ativar: dados não bateram: %+v", posAtivar)
	}
}

func TestVeiculoCadastrar_PlacaDuplicada_Retorna409(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	var primeiro httpveiculo.VeiculoResponse
	rec := doRequest(t, http.MethodPost, "/v1/veiculos", loginAdmin.AccessToken,
		httpveiculo.CadastrarVeiculoRequest{Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 15000, Ano: 2020, Cor: "Prata"},
		&primeiro)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup (primeiro cadastro) falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPost, "/v1/veiculos", loginAdmin.AccessToken,
		httpveiculo.CadastrarVeiculoRequest{Placa: "abc-1d23", Marca: "Outra", Modelo: "Outro", QuilometragemAtual: 0, Ano: 2021, Cor: "Verde"},
		nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("Cadastrar com placa duplicada deveria ser 409, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestVeiculoConsultarPorPlaca_NaoEncontrado_Retorna404(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	rec := doRequest(t, http.MethodGet, "/v1/veiculos/placa/ZZZ9Z99", loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("ConsultarPorPlaca inexistente deveria ser 404, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestVeiculoListar_ComPaginacaoEFiltro_Retorna200(t *testing.T) {
	resetDB(t)

	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)

	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	veiculos := []httpveiculo.CadastrarVeiculoRequest{
		{
			Placa:              "ABC1D23",
			Marca:              "Fiat",
			Modelo:             "Uno",
			QuilometragemAtual: 15000,
			Ano:                2020,
			Cor:                "Prata",
		},
		{
			Placa:              "DEF4G56",
			Marca:              "Fiat",
			Modelo:             "Argo",
			QuilometragemAtual: 20000,
			Ano:                2022,
			Cor:                "Preto",
		},
		{
			Placa:              "GHI7J89",
			Marca:              "Volkswagen",
			Modelo:             "Gol",
			QuilometragemAtual: 30000,
			Ano:                2021,
			Cor:                "Branco",
		},
	}

	for _, req := range veiculos {
		rec := doRequest(t, http.MethodPost, "/v1/veiculos", loginAdmin.AccessToken, req, nil)

		if rec.Code != http.StatusCreated {
			t.Fatalf("setup cadastro veículo falhou: status %d, body %q", rec.Code, rec.Body.String())
		}
	}

	var resp httpveiculo.ListarVeiculosResponse

	rec := doRequest(
		t,
		http.MethodGet,
		"/v1/veiculos?marca=Fiat&limit=1&offset=0&order=ano&direction=DESC",
		loginAdmin.AccessToken,
		nil,
		&resp,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("Listar: status %d, body %q", rec.Code, rec.Body.String())
	}

	if resp.Total != 2 {
		t.Fatalf("Listar: total esperado 2, veio %d", resp.Total)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("Listar: esperado 1 item, vieram %d", len(resp.Items))
	}

	if resp.Items[0].Marca != "Fiat" {
		t.Fatalf("Listar: marca esperada Fiat, veio %q", resp.Items[0].Marca)
	}

	if resp.Items[0].Ano != 2022 {
		t.Fatalf("Listar: esperado veículo mais recente primeiro, veio ano %d", resp.Items[0].Ano)
	}

	if resp.Offset != 0 || resp.Limit != 1 || resp.Order != "ano" || resp.Direction != "DESC" {
		t.Fatalf("Listar: metadados inesperados: %+v", resp)
	}
}

func TestVeiculoListar_SemRegistros_RetornaPaginaVazia(t *testing.T) {
	resetDB(t)

	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)

	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	var resp httpveiculo.ListarVeiculosResponse

	rec := doRequest(t, http.MethodGet, "/v1/veiculos", loginAdmin.AccessToken, nil, &resp)

	if rec.Code != http.StatusOK {
		t.Fatalf("Listar vazio: status %d, body %q", rec.Code, rec.Body.String())
	}

	if resp.Total != 0 || len(resp.Items) != 0 {
		t.Fatalf("Listar vazio: resposta inesperada: %+v", resp)
	}
}
