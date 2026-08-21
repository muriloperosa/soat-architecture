//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httppeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/peca"
)

// TestPecaLifecycle_TodosOsEndpoints exercita, em sequência, os 6 endpoints
// de /v1/pecas (Cadastrar, ConsultarPorID, Atualizar, ReporEstoque, Inativar,
// Ativar) sobre a mesma peça, contra banco real.
func TestPecaLifecycle_TodosOsEndpoints(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	// 1. Cadastrar — admin cadastra peça, criado_por vem do token, não do body
	var criada httppeca.PecaResponse
	rec := doRequest(t, http.MethodPost, "/v1/pecas", loginAdmin.AccessToken,
		httppeca.CadastrarPecaRequest{Nome: "Pastilha de freio", Marca: "Bosch", Descricao: "Pastilha dianteira", Preco: 89.9, QuantidadeEmEstoque: 20, EstoqueMinimo: 5},
		&criada)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Cadastrar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if criada.Codigo == "" || !criada.Ativo || criada.CriadoPor == 0 {
		t.Fatalf("Cadastrar: resposta inesperada: %+v", criada)
	}

	// 2. ConsultarPorID, confirma os dados persistidos
	var consultada httppeca.PecaResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/pecas/%d", criada.ID), loginAdmin.AccessToken, nil, &consultada)
	if rec.Code != http.StatusOK || consultada.ID != criada.ID || consultada.QuantidadeEmEstoque != 20 {
		t.Fatalf("ConsultarPorID: status %d, body %+v", rec.Code, consultada)
	}

	// 3. Atualizar, dados cadastrais e estoque mínimo mudam, estoque físico não
	var atualizada httppeca.PecaResponse
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/v1/pecas/%d", criada.ID), loginAdmin.AccessToken,
		httppeca.AtualizarPecaRequest{Nome: "Pastilha de freio dianteira", Marca: "Bosch", Descricao: "Pastilha dianteira reforcada", Preco: 99.9, EstoqueMinimo: 8},
		&atualizada)
	if rec.Code != http.StatusOK {
		t.Fatalf("Atualizar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if atualizada.Nome != "Pastilha de freio dianteira" || atualizada.EstoqueMinimo != 8 || atualizada.QuantidadeEmEstoque != 20 {
		t.Fatalf("Atualizar: dados não bateram: %+v", atualizada)
	}

	// 4. ReporEstoque, soma quantidade ao estoque físico
	var reposta httppeca.PecaResponse
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/pecas/%d/repor-estoque", criada.ID), loginAdmin.AccessToken,
		httppeca.ReporEstoqueRequest{Quantidade: 10}, &reposta)
	if rec.Code != http.StatusOK || reposta.QuantidadeEmEstoque != 30 {
		t.Fatalf("ReporEstoque: status %d, body %+v", rec.Code, reposta)
	}

	// 5. Inativar
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/pecas/%d/inativar", criada.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Inativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	var posInativar httppeca.PecaResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/pecas/%d", criada.ID), loginAdmin.AccessToken, nil, &posInativar)
	if rec.Code != http.StatusOK || posInativar.Ativo {
		t.Fatalf("ConsultarPorID pós-Inativar deveria vir ativo=false: %+v", posInativar)
	}

	// 6. Ativar, reverte a inativação
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/pecas/%d/ativar", criada.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Ativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	var posAtivar httppeca.PecaResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/pecas/%d", criada.ID), loginAdmin.AccessToken, nil, &posAtivar)
	if rec.Code != http.StatusOK || !posAtivar.Ativo || posAtivar.QuantidadeEmEstoque != 30 {
		t.Fatalf("ConsultarPorID pós-Ativar: dados não bateram: %+v", posAtivar)
	}
}
