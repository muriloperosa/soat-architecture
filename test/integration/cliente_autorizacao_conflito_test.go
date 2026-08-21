//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpcliente "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/cliente"
)

func TestClienteRoutes_SemToken_Retorna401(t *testing.T) {
	resetDB(t)
	rec := doRequest(t, http.MethodPost, "/v1/clientes", "",
		httpcliente.CriarClienteRequest{
			Nome: "Maria Silva", Email: "maria@email.com", Senha: "senhaInicial123",
			Documento: "52998224725", TipoPessoa: "PF", Telefone: "11999998888",
		}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("criação sem token deveria ser 401, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestClienteRoutes_TokenClienteNaoAcessaGestaoInterna(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	var criado httpcliente.ClienteResponse
	rec := doRequest(t, http.MethodPost, "/v1/clientes", loginAdmin.AccessToken,
		httpcliente.CriarClienteRequest{
			Nome: "Maria Silva", Email: "maria@email.com", Senha: "senhaInicial123",
			Documento: "52998224725", TipoPessoa: "PF", Telefone: "11999998888",
		}, &criado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("setup cliente: status %d, body %q", rec.Code, rec.Body.String())
	}

	loginCliente := doLoginCliente(t, "maria@email.com", "senhaInicial123")
	rec = doRequest(t, http.MethodGet, "/v1/clientes/"+fmt.Sprint(criado.ID), loginCliente.AccessToken, nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("token de cliente em rota interna deveria ser 403, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestCriarCliente_DocumentoDuplicado_Retorna409(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	primeiro := httpcliente.CriarClienteRequest{
		Nome: "Maria Silva", Email: "maria@email.com", Senha: "senhaInicial123",
		Documento: "52998224725", TipoPessoa: "PF", Telefone: "11999998888",
	}
	rec := doRequest(t, http.MethodPost, "/v1/clientes", loginAdmin.AccessToken, primeiro, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("primeira criação falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	duplicado := primeiro
	duplicado.Email = "outro@email.com"
	rec = doRequest(t, http.MethodPost, "/v1/clientes", loginAdmin.AccessToken, duplicado, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("documento duplicado deveria ser 409, veio %d, body %q", rec.Code, rec.Body.String())
	}

	var count int64
	if err := testDB.Table("clientes").Where("documento = ?", "52998224725").Count(&count).Error; err != nil {
		t.Fatalf("erro ao contar clientes: %v", err)
	}
	if count != 1 {
		t.Fatalf("esperava exatamente 1 cliente com o documento, encontrou %d", count)
	}
}
