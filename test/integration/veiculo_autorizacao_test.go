//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpveiculo "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/veiculo"
)

func TestAutorizacaoVeiculo_NaoAdminNaoCadastraMasConsulta(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	seedUsuario(t, "Bia Lima", "bia@oficina.com", "senha123", shared.PapelMecanico)

	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	var criado httpveiculo.VeiculoResponse
	rec := doRequest(t, http.MethodPost, "/v1/veiculos", loginAdmin.AccessToken,
		httpveiculo.CadastrarVeiculoRequest{Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 15000, Ano: 2020, Cor: "Prata"},
		&criado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup (Cadastrar por admin) falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	loginMecanico := doLogin(t, "bia@oficina.com", "senha123")

	rec = doRequest(t, http.MethodPost, "/v1/veiculos", loginMecanico.AccessToken,
		httpveiculo.CadastrarVeiculoRequest{Placa: "XYZ9W88", Marca: "Marca", Modelo: "Modelo", QuilometragemAtual: 0, Ano: 2021, Cor: "Branco"},
		nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /v1/veiculos por não-admin deveria ser 403, veio %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/veiculos/%d", criado.ID), loginMecanico.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/veiculos/%d por mecânico deveria ser 200, veio %d, body %q", criado.ID, rec.Code, rec.Body.String())
	}
}
