//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
	httpusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/usuario"
)

// TestSenhaProvisoria_RoundTrip prova que a regra "senha definida pelo
// admin nasce provisória, força troca no primeiro acesso" sobrevive à ida e
// volta pelo GORM/enum do banco.
func TestSenhaProvisoria_RoundTrip(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	// admin cria usuário novo via HTTP (não via seedUsuario, pra exercitar o
	// handler real e o requer_alterar_senha na resposta de criação)
	var criado httpusuario.UsuarioResponse
	rec := doRequest(t, http.MethodPost, "/v1/usuarios", loginAdmin.AccessToken,
		httpusuario.CriarUsuarioRequest{Nome: "Bia Lima", Email: "bia@oficina.com", Senha: "senhaInicial123", Papel: shared.PapelMecanico},
		&criado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /v1/usuarios falhou: status %d, body %q", rec.Code, rec.Body.String())
	}
	if !criado.RequerAlterarSenha {
		t.Fatal("usuário recém-criado deveria nascer com requer_alterar_senha=true")
	}

	// login do usuário novo confirma requer_alterar_senha=true também na
	// resposta de login (reidratado do banco, não é o mesmo objeto em memória)
	loginNovo := doLogin(t, "bia@oficina.com", "senhaInicial123")
	if !loginNovo.RequerAlterarSenha {
		t.Fatal("login do usuário novo deveria retornar requer_alterar_senha=true")
	}

	// troca a senha (self-service)
	rec = doRequest(t, http.MethodPut, "/v1/usuarios/me/senha", loginNovo.AccessToken,
		httpusuario.AlterarSenhaRequest{SenhaNova: "senhaDefinitiva123"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PUT /me/senha falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	// senha antiga não funciona mais
	rec = doRequest(t, http.MethodPost, "/v1/auth/login", "", httpauth.LoginRequest{Email: "bia@oficina.com", Senha: "senhaInicial123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login com senha antiga deveria ser 401, veio %d", rec.Code)
	}

	// login com a senha nova confirma requer_alterar_senha=false, a flag
	// persistiu no banco, não só na entidade em memória do request anterior
	loginFinal := doLogin(t, "bia@oficina.com", "senhaDefinitiva123")
	if loginFinal.RequerAlterarSenha {
		t.Fatal("login pós-troca de senha deveria retornar requer_alterar_senha=false")
	}
}
