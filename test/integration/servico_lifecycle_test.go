//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/servico"
)

func floatPtr(v float64) *float64 { return &v }

// TestServicoLifecycle_TodosOsEndpoints exercita, em sequência, os 6
// endpoints de /v1/servicos (Criar, Listar, Buscar, Atualizar, Inativar,
// Ativar) sobre o mesmo serviço, contra banco real.
func TestServicoLifecycle_TodosOsEndpoints(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	// 1. Criar — criado_por vem do token, não do body
	var criado httpservico.ServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/servicos", loginAdmin.AccessToken,
		httpservico.CriarServicoRequest{Nome: "Troca de óleo", Descricao: "Troca de óleo e filtro", PrecoBase: floatPtr(150.5), TempoEstimadoMinutos: 60},
		&criado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Criar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if criado.ID == 0 || !criado.Ativo || criado.CriadoPor == 0 {
		t.Fatalf("Criar: resposta inesperada: %+v", criado)
	}

	// 2. Listar — confirma que o serviço criado aparece na listagem
	var listagem httpservico.ListarServicosResponse
	rec = doRequest(t, http.MethodGet, "/v1/servicos", loginAdmin.AccessToken, nil, &listagem)
	if rec.Code != http.StatusOK {
		t.Fatalf("Listar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if listagem.Total != 1 || len(listagem.Items) != 1 || listagem.Items[0].ID != criado.ID {
		t.Fatalf("Listar: resposta inesperada: %+v", listagem)
	}

	// 3. Buscar por ID, confirma os dados persistidos
	var consultado httpservico.ServicoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/servicos/%d", criado.ID), loginAdmin.AccessToken, nil, &consultado)
	if rec.Code != http.StatusOK || consultado.ID != criado.ID || consultado.TempoEstimadoMinutos != 60 {
		t.Fatalf("Buscar: status %d, body %+v", rec.Code, consultado)
	}

	// 4. Atualizar
	var atualizado httpservico.ServicoResponse
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/v1/servicos/%d", criado.ID), loginAdmin.AccessToken,
		httpservico.AtualizarServicoRequest{Nome: "Troca de óleo sintético", Descricao: "Troca de óleo sintético e filtro", PrecoBase: floatPtr(180.0), TempoEstimadoMinutos: 45},
		&atualizado)
	if rec.Code != http.StatusOK {
		t.Fatalf("Atualizar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if atualizado.Nome != "Troca de óleo sintético" || atualizado.PrecoBase != 180.0 || atualizado.TempoEstimadoMinutos != 45 {
		t.Fatalf("Atualizar: dados não bateram: %+v", atualizado)
	}

	// 5. Inativar
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/servicos/%d/inativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Inativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	var posInativar httpservico.ServicoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/servicos/%d", criado.ID), loginAdmin.AccessToken, nil, &posInativar)
	if rec.Code != http.StatusOK || posInativar.Ativo {
		t.Fatalf("Buscar pós-Inativar deveria vir ativo=false: %+v", posInativar)
	}

	// 6. Ativar, reverte a inativação
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/servicos/%d/ativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Ativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	var posAtivar httpservico.ServicoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/servicos/%d", criado.ID), loginAdmin.AccessToken, nil, &posAtivar)
	if rec.Code != http.StatusOK || !posAtivar.Ativo {
		t.Fatalf("Buscar pós-Ativar: dados não bateram: %+v", posAtivar)
	}
}

// TestServicoListar_ComPaginacaoOrdenacaoEFiltro_Retorna200 confirma que a
// listagem de serviços aceita filtro por nome e ordenação por preço.
func TestServicoListar_ComPaginacaoOrdenacaoEFiltro_Retorna200(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	criar := func(nome string, preco float64) {
		t.Helper()
		rec := doRequest(t, http.MethodPost, "/v1/servicos", loginAdmin.AccessToken,
			httpservico.CriarServicoRequest{Nome: nome, Descricao: "Serviço de teste", PrecoBase: floatPtr(preco), TempoEstimadoMinutos: 30},
			nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("Criar %s: status %d, body %q", nome, rec.Code, rec.Body.String())
		}
	}

	criar("Troca de óleo", 150.0)
	criar("Troca de óleo sintético", 220.0)
	criar("Alinhamento", 90.0)

	var resp httpservico.ListarServicosResponse
	rec := doRequest(t, http.MethodGet, "/v1/servicos?nome=óleo&page=1&order=preco_base&direction=DESC",
		loginAdmin.AccessToken, nil, &resp)
	if rec.Code != http.StatusOK {
		t.Fatalf("Listar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if resp.Total != 2 || len(resp.Items) != 2 {
		t.Fatalf("Listar: esperado total=2, recebido %+v", resp)
	}
	if resp.Items[0].PrecoBase != 220.0 || resp.Items[1].PrecoBase != 150.0 {
		t.Fatalf("Listar: ordenação por preco_base DESC inesperada: %+v", resp.Items)
	}
	if resp.Direction != "DESC" || resp.Order != "preco_base" {
		t.Fatalf("Listar: metadados de ordenação inesperados: %+v", resp)
	}
}
