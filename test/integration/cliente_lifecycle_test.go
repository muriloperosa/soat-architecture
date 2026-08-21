//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
	httpcliente "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/cliente"
)

// TestClienteLifecycle_TodosOsEndpoints exercita o fluxo completo de cliente
// usando router, middlewares, wiring, casos de uso, GORM e MySQL reais.
func TestClienteLifecycle_TodosOsEndpoints(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	// 1. O usuário interno cria o cliente. criado_por deve vir do subject do
	// token, nunca do corpo enviado pelo consumidor.
	var criado httpcliente.ClienteResponse
	rec := doRequest(t, http.MethodPost, "/v1/clientes", loginAdmin.AccessToken,
		httpcliente.CriarClienteRequest{
			Nome: "Maria Silva", Email: "maria@email.com", Senha: "senhaInicial123",
			Documento: "52998224725", TipoPessoa: "PF", Telefone: "11999998888",
		}, &criado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Criar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if criado.ID == 0 || !criado.Ativo || !criado.RequerAlterarSenha || criado.CriadoPor != admin.ID {
		t.Fatalf("Criar: resposta inesperada: %+v", criado)
	}

	// 2. O mesmo usuário interno consulta por ID e por documento.
	var porID httpcliente.ClienteResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/clientes/%d", criado.ID), loginAdmin.AccessToken, nil, &porID)
	if rec.Code != http.StatusOK || porID.ID != criado.ID {
		t.Fatalf("BuscarPorID: status %d, body %+v", rec.Code, porID)
	}

	var porDocumento httpcliente.ClienteResponse
	rec = doRequest(t, http.MethodGet, "/v1/clientes/documento/52998224725", loginAdmin.AccessToken, nil, &porDocumento)
	if rec.Code != http.StatusOK || porDocumento.ID != criado.ID {
		t.Fatalf("BuscarPorDocumento: status %d, body %+v", rec.Code, porDocumento)
	}

	// 3. Atualiza os dados sem alterar documento, criador ou senha.
	var atualizado httpcliente.ClienteResponse
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/v1/clientes/%d", criado.ID), loginAdmin.AccessToken,
		httpcliente.AtualizarClienteRequest{
			Nome: "Maria S. Souza", Email: "maria.souza@email.com", Telefone: "11988887777",
		}, &atualizado)
	if rec.Code != http.StatusOK {
		t.Fatalf("Atualizar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if atualizado.Nome != "Maria S. Souza" || atualizado.Email != "maria.souza@email.com" || atualizado.CriadoPor != admin.ID {
		t.Fatalf("Atualizar: dados inesperados: %+v", atualizado)
	}

	// O e-mail anterior deixa de autenticar e o novo mantém a senha inicial.
	rec = doRequest(t, http.MethodPost, "/v1/auth/cliente/login", "",
		httpauth.LoginRequest{Email: "maria@email.com", Senha: "senhaInicial123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login com email antigo deveria ser 401, veio %d", rec.Code)
	}

	loginCliente := doLoginCliente(t, "maria.souza@email.com", "senhaInicial123")
	if !loginCliente.RequerAlterarSenha {
		t.Fatal("cliente recém-criado deveria exigir alteração de senha")
	}

	// 4. O cliente altera a própria senha usando o subject do token.
	rec = doRequest(t, http.MethodPut, "/v1/clientes/me/senha", loginCliente.AccessToken,
		httpcliente.AlterarSenhaRequest{SenhaNova: "senhaDefinitiva123"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("AlterarSenha: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPost, "/v1/auth/cliente/login", "",
		httpauth.LoginRequest{Email: "maria.souza@email.com", Senha: "senhaInicial123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login com senha antiga deveria ser 401, veio %d", rec.Code)
	}
	loginDefinitivo := doLoginCliente(t, "maria.souza@email.com", "senhaDefinitiva123")
	if loginDefinitivo.RequerAlterarSenha {
		t.Fatal("login após troca deveria retornar requer_alterar_senha=false")
	}

	// 5. Inativar bloqueia login e Ativar permite autenticar novamente.
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/clientes/%d/inativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Inativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, http.MethodPost, "/v1/auth/cliente/login", "",
		httpauth.LoginRequest{Email: "maria.souza@email.com", Senha: "senhaDefinitiva123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login de cliente inativo deveria ser 401, veio %d", rec.Code)
	}

	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/clientes/%d/ativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Ativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if doLoginCliente(t, "maria.souza@email.com", "senhaDefinitiva123").AccessToken == "" {
		t.Fatal("login após reativação deveria emitir access token")
	}
}
