//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpcliente "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/cliente"
)

// TestClienteInativacaoMidSessao_TokenAntigoParaDeFuncionar prova que o
// ClienteStatusRepo é consultado em cada request e não apenas durante login.
func TestClienteInativacaoMidSessao_TokenAntigoParaDeFuncionar(t *testing.T) {
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

	// O token do cliente funciona antes da inativação.
	rec = doRequest(t, http.MethodPut, "/v1/clientes/me/senha", loginCliente.AccessToken,
		httpcliente.AlterarSenhaRequest{SenhaNova: "senhaDefinitiva123"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("token antes da inativação deveria funcionar: status %d", rec.Code)
	}

	// Obtém um token novo e então inativa o cliente por uma rota interna.
	loginCliente = doLoginCliente(t, "maria@email.com", "senhaDefinitiva123")
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/clientes/%d/inativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("inativação falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	// O mesmo access token, ainda dentro do TTL, passa a ser rejeitado.
	rec = doRequest(t, http.MethodPut, "/v1/clientes/me/senha", loginCliente.AccessToken,
		httpcliente.AlterarSenhaRequest{SenhaNova: "outraSenha123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("token antigo após inativação deveria ser 401, veio %d, body %q", rec.Code, rec.Body.String())
	}
}
