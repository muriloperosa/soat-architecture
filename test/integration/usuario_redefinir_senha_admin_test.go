//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
	httpusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/usuario"
)

// TestAtualizarUsuario_ComSenhaNova_ForcaTrocaNoProximoLogin prova
// RedefinirSenha via PUT /:id (admin reseta a senha de outro usuário),
// diferente de AlterarSenha (self-service), que destrava o estado
// provisório, aqui o admin *orça o estado provisório de volta, mesmo pra
// um usuário que já tinha trocado a senha antes. É o caminho "esqueci minha
// senha, o admin redefiniu pra mim" nenhum outro cenário cobre isso.
func TestAtualizarUsuario_ComSenhaNova_ForcaTrocaNoProximoLogin(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	alvo := seedUsuario(t, "Bia Lima", "bia@oficina.com", "senhaInicial123", shared.PapelMecanico)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	// alvo já destravou o próprio primeiro acesso antes, prova que o reset
	// do admin força requer_alterar_senha=true de novo, não é só o valor
	// inicial de NewUsuario nunca tendo sido trocado
	loginAlvoInicial := doLogin(t, "bia@oficina.com", "senhaInicial123")
	if !loginAlvoInicial.RequerAlterarSenha {
		t.Fatal("setup: usuário deveria nascer com requer_alterar_senha=true")
	}
	rec := doRequest(t, http.MethodPut, "/v1/usuarios/me/senha", loginAlvoInicial.AccessToken,
		httpusuario.AlterarSenhaRequest{SenhaNova: "senhaPropria123"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("setup AlterarSenha: status %d, body %q", rec.Code, rec.Body.String())
	}
	loginAlvoPosTroca := doLogin(t, "bia@oficina.com", "senhaPropria123")
	if loginAlvoPosTroca.RequerAlterarSenha {
		t.Fatal("setup: após trocar a própria senha, requer_alterar_senha deveria ser false")
	}

	// admin redefine a senha do alvo via PUT /:id (não é o alvo trocando a
	// própria senha, é o admin agindo em nome dele)
	var atualizado httpusuario.UsuarioResponse
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/v1/usuarios/%d", alvo.ID), loginAdmin.AccessToken,
		httpusuario.AtualizarUsuarioRequest{Nome: "Bia Lima", Email: "bia@oficina.com", SenhaNova: "senhaResetadaPeloAdmin123", Papel: shared.PapelMecanico},
		&atualizado)
	if rec.Code != http.StatusOK {
		t.Fatalf("Atualizar com senha_nova: status %d, body %q", rec.Code, rec.Body.String())
	}
	if !atualizado.RequerAlterarSenha {
		t.Fatal("resposta do Atualizar deveria retornar requer_alterar_senha=true após reset do admin")
	}

	// senha antiga (a que o alvo tinha escolhido) já não funciona mais
	rec = doRequest(t, http.MethodPost, "/v1/auth/login", "", httpauth.LoginRequest{Email: "bia@oficina.com", Senha: "senhaPropria123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login com senha antiga pós-reset deveria ser 401, veio %d", rec.Code)
	}

	// login com a senha que o admin definiu funciona e força troca de novo
	loginPosReset := doLogin(t, "bia@oficina.com", "senhaResetadaPeloAdmin123")
	if !loginPosReset.RequerAlterarSenha {
		t.Fatal("login com senha resetada pelo admin deveria retornar requer_alterar_senha=true")
	}

	// e o alvo consegue destravar de novo, mesmo fluxo de sempre
	rec = doRequest(t, http.MethodPut, "/v1/usuarios/me/senha", loginPosReset.AccessToken,
		httpusuario.AlterarSenhaRequest{SenhaNova: "senhaFinalDoAlvo123"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("AlterarSenha final: status %d, body %q", rec.Code, rec.Body.String())
	}
	loginFinal := doLogin(t, "bia@oficina.com", "senhaFinalDoAlvo123")
	if loginFinal.RequerAlterarSenha {
		t.Fatal("login final deveria retornar requer_alterar_senha=false")
	}
}
