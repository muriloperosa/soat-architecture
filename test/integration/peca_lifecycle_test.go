//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httppeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/peca"
)

// TestPecaLifecycle_TodosOsEndpoints exercita, em sequência, os 7 endpoints
// de /v1/pecas (Cadastrar, Listar, ConsultarPorID, Atualizar, ReporEstoque,
// Inativar, Ativar) sobre a mesma peça, contra banco real.
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

	// 2. Listar — confirma que a peça cadastrada aparece na listagem
	var listagem httppeca.ListarPecasResponse

	rec = doRequest(t, http.MethodGet, "/v1/pecas", loginAdmin.AccessToken, nil, &listagem)

	if rec.Code != http.StatusOK {
		t.Fatalf("Listar: status %d, body %q", rec.Code, rec.Body.String())
	}

	if listagem.Total != 1 {
		t.Fatalf("Listar: esperado total=1, recebido %d", listagem.Total)
	}

	if len(listagem.Items) != 1 {
		t.Fatalf("Listar: esperado 1 item, recebido %d", len(listagem.Items))
	}

	if listagem.Items[0].ID != criada.ID {
		t.Fatalf("Listar: esperado ID %d, recebido %d", criada.ID, listagem.Items[0].ID)
	}

	if listagem.Offset != 0 || listagem.Limit != 20 || listagem.Order != "id" || listagem.Direction != "ASC" {
		t.Fatalf("Listar: paginação padrão inesperada: %+v", listagem)
	}

	// 3. ConsultarPorID, confirma os dados persistidos
	var consultada httppeca.PecaResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/pecas/%d", criada.ID), loginAdmin.AccessToken, nil, &consultada)
	if rec.Code != http.StatusOK || consultada.ID != criada.ID || consultada.QuantidadeEmEstoque != 20 {
		t.Fatalf("ConsultarPorID: status %d, body %+v", rec.Code, consultada)
	}

	// 4. Atualizar, dados cadastrais e estoque mínimo mudam, estoque físico não
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

	// 5. ReporEstoque, soma quantidade ao estoque físico
	var reposta httppeca.PecaResponse
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/pecas/%d/repor-estoque", criada.ID), loginAdmin.AccessToken,
		httppeca.ReporEstoqueRequest{Quantidade: 10}, &reposta)
	if rec.Code != http.StatusOK || reposta.QuantidadeEmEstoque != 30 {
		t.Fatalf("ReporEstoque: status %d, body %+v", rec.Code, reposta)
	}

	// 6. Inativar
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/pecas/%d/inativar", criada.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Inativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	var posInativar httppeca.PecaResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/pecas/%d", criada.ID), loginAdmin.AccessToken, nil, &posInativar)
	if rec.Code != http.StatusOK || posInativar.Ativo {
		t.Fatalf("ConsultarPorID pós-Inativar deveria vir ativo=false: %+v", posInativar)
	}

	// 7. Ativar, reverte a inativação
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

func TestPecaListar_ComPaginacaoOrdenacaoEFiltro_Retorna200(t *testing.T) {
	resetDB(t)

	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)

	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	cadastrar := func(nome string, marca string, preco float64, quantidade int, estoqueMinimo int) httppeca.PecaResponse {
		t.Helper()

		var criada httppeca.PecaResponse

		rec := doRequest(
			t,
			http.MethodPost,
			"/v1/pecas",
			loginAdmin.AccessToken,
			httppeca.CadastrarPecaRequest{
				Nome:                nome,
				Marca:               marca,
				Descricao:           "Peça para teste",
				Preco:               preco,
				QuantidadeEmEstoque: quantidade,
				EstoqueMinimo:       estoqueMinimo,
			},
			&criada,
		)

		if rec.Code != http.StatusCreated {
			t.Fatalf("Cadastrar %s: status %d, body %q", nome, rec.Code, rec.Body.String())
		}

		return criada
	}

	cadastrar("Pastilha Bosch Premium", "Bosch", 150.00, 20, 5)

	cadastrar("Pastilha Bosch Standard", "Bosch", 90.00, 15, 5)

	cadastrar("Filtro de óleo", "Mann", 50.00, 30, 10)

	var resp httppeca.ListarPecasResponse

	rec := doRequest(
		t,
		http.MethodGet,
		"/v1/pecas?marca=Bosch&limit=1&offset=0&order=preco&direction=DESC",
		loginAdmin.AccessToken,
		nil,
		&resp,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("Listar: status %d, body %q", rec.Code, rec.Body.String())
	}

	if resp.Total != 2 {
		t.Fatalf("Listar: esperado total=2, recebido %d", resp.Total)
	}

	if len(resp.Items) != 1 {
		t.Fatalf("Listar: esperado 1 item, recebido %d", len(resp.Items))
	}

	item := resp.Items[0]

	if item.Marca != "Bosch" {
		t.Fatalf("Listar: esperado marca Bosch, recebido %q", item.Marca)
	}

	if item.Preco != 150.00 {
		t.Fatalf("Listar: esperado maior preço primeiro (150), recebido %.2f", item.Preco)
	}

	if resp.Offset != 0 {
		t.Fatalf("Listar: esperado offset=0, recebido %d", resp.Offset)
	}

	if resp.Limit != 1 {
		t.Fatalf("Listar: esperado limit=1, recebido %d", resp.Limit)
	}

	if resp.Order != "preco" {
		t.Fatalf("Listar: esperado order=preco, recebido %q", resp.Order)
	}

	if resp.Direction != "DESC" {
		t.Fatalf("Listar: esperado direction=DESC, recebido %q", resp.Direction)
	}
}
