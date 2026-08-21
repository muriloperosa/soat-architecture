//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httppeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/peca"
)

// TestAutorizacaoPeca_NaoAdminNaoCadastraMasConsulta prova, contra o banco
// de verdade, a diferença de papel desenhada pras rotas de /v1/pecas: um
// mecânico autenticado não pode cadastrar (403, rota restrita a admin), mas
// pode consultar (200, qualquer usuário interno).
func TestAutorizacaoPeca_NaoAdminNaoCadastraMasConsulta(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	seedUsuario(t, "Bia Lima", "bia@oficina.com", "senha123", shared.PapelMecanico)

	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	var criada httppeca.PecaResponse
	rec := doRequest(t, http.MethodPost, "/v1/pecas", loginAdmin.AccessToken,
		httppeca.CadastrarPecaRequest{Nome: "Pastilha de freio", Marca: "Bosch", Descricao: "Pastilha dianteira", Preco: 89.9, QuantidadeEmEstoque: 20, EstoqueMinimo: 5},
		&criada)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup (Cadastrar por admin) falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	loginMecanico := doLogin(t, "bia@oficina.com", "senha123")

	rec = doRequest(t, http.MethodPost, "/v1/pecas", loginMecanico.AccessToken,
		httppeca.CadastrarPecaRequest{Nome: "Outra peça", Marca: "Marca", Descricao: "Descricao", Preco: 10, QuantidadeEmEstoque: 1, EstoqueMinimo: 0},
		nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /v1/pecas por não-admin deveria ser 403, veio %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/pecas/%d", criada.ID), loginMecanico.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /v1/pecas/%d por mecânico deveria ser 200, veio %d, body %q", criada.ID, rec.Code, rec.Body.String())
	}
}
