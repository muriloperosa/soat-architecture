//go:build integration

package integration_test

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpcliente "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/cliente"
)

func TestClienteListagem_PaginaOrdenaEFiltra(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	clientes := []httpcliente.CriarClienteRequest{
		{
			Nome: "Maria Alves", Email: "maria.alves@email.com", Senha: "senha123",
			Documento: "52998224725", TipoPessoa: "PF", Telefone: "11999990001",
		},
		{
			Nome: "Bruno Lima", Email: "bruno@email.com", Senha: "senha123",
			Documento: "12345678909", TipoPessoa: "PF", Telefone: "11999990002",
		},
		{
			Nome: "Maria Souza", Email: "maria.souza@email.com", Senha: "senha123",
			Documento: "10000000280", TipoPessoa: "PF", Telefone: "11999990003",
		},
	}
	for _, cliente := range clientes {
		rec := doRequest(t, http.MethodPost, "/v1/clientes", login.AccessToken, cliente, nil)
		if rec.Code != http.StatusCreated {
			t.Fatalf("criação de cliente falhou: status %d, body %q", rec.Code, rec.Body.String())
		}
	}

	values := url.Values{}
	values.Set("offset", "1")
	values.Set("limit", "1")
	values.Set("order", "nome")
	values.Set("direction", "DESC")
	values.Set("nome", "Maria")
	values.Set("ativo", "true")

	var response httpcliente.ListarClientesResponse
	rec := doRequest(
		t,
		http.MethodGet,
		"/v1/clientes?"+values.Encode(),
		login.AccessToken,
		nil,
		&response,
	)

	if rec.Code != http.StatusOK {
		t.Fatalf("listagem falhou: status %d, body %q", rec.Code, rec.Body.String())
	}
	if response.Total != 2 || len(response.Items) != 1 {
		t.Fatalf("paginação inesperada: %+v", response)
	}
	if response.Items[0].Nome != "Maria Alves" {
		t.Fatalf("ordenação/offset inesperados: %+v", response.Items)
	}
	if response.Offset != 1 || response.Limit != 1 || response.Order != "nome" || response.Direction != "DESC" {
		t.Fatalf("metadados inesperados: %+v", response)
	}
}

func TestClienteListagem_RejeitaFiltroNaoPermitido(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	values := url.Values{}
	values.Set("senha_hash", "segredo")
	rec := doRequest(
		t,
		http.MethodGet,
		"/v1/clientes?"+values.Encode(),
		login.AccessToken,
		nil,
		nil,
	)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("filtro não permitido deveria retornar 400, veio %d: %q", rec.Code, rec.Body.String())
	}
}
