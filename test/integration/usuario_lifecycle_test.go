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

// TestUsuarioLifecycle_TodosOsEndpoints exercita, em sequência, os 6
// endpoints de /v1/usuarios (Criar, Me, AlterarSenha, Atualizar, Inativar,
// Ativar) sobre o mesmo usuário, os outros testes de integração cobrem
// Criar/Me/AlterarSenha/Inativar isoladamente ou em combinação parcial;
// este é o único que passa por Atualizar e Ativar, e o único que confirma
// que Ativar reverte Inativar de ponta a ponta (login volta a funcionar).
func TestUsuarioLifecycle_TodosOsEndpoints(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	// 1. Criar — admin cria usuário novo, senha nasce provisória
	var criado httpusuario.UsuarioResponse
	rec := doRequest(t, http.MethodPost, "/v1/usuarios", loginAdmin.AccessToken,
		httpusuario.CriarUsuarioRequest{Nome: "Bia Lima", Email: "bia@oficina.com", Senha: "senhaInicial123", Papel: string(shared.PapelMecanico)},
		&criado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("Criar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if !criado.RequerAlterarSenha || !criado.Ativo || criado.Papel != string(shared.PapelMecanico) {
		t.Fatalf("Criar: resposta inesperada: %+v", criado)
	}

	// 2. Login do usuário novo + Me, confirma dados e requer_alterar_senha=true
	loginBia := doLogin(t, "bia@oficina.com", "senhaInicial123")
	if !loginBia.RequerAlterarSenha {
		t.Fatal("login inicial deveria retornar requer_alterar_senha=true")
	}
	var me httpusuario.UsuarioResponse
	rec = doRequest(t, http.MethodGet, "/v1/usuarios/me", loginBia.AccessToken, nil, &me)
	if rec.Code != http.StatusOK || me.ID != criado.ID || me.Nome != "Bia Lima" {
		t.Fatalf("Me: status %d, body %+v", rec.Code, me)
	}

	// 3. AlterarSenha (self-service), destrava o primeiro acesso
	rec = doRequest(t, http.MethodPut, "/v1/usuarios/me/senha", loginBia.AccessToken,
		httpusuario.AlterarSenhaRequest{SenhaNova: "senhaDefinitiva123"}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("AlterarSenha: status %d, body %q", rec.Code, rec.Body.String())
	}
	loginBiaPosTroca := doLogin(t, "bia@oficina.com", "senhaDefinitiva123")
	if loginBiaPosTroca.RequerAlterarSenha {
		t.Fatal("login pós-troca deveria retornar requer_alterar_senha=false")
	}

	// 4. Atualizar, admin troca nome, email e papel do usuário
	var atualizado httpusuario.UsuarioResponse
	rec = doRequest(t, http.MethodPut, fmt.Sprintf("/v1/usuarios/%d", criado.ID), loginAdmin.AccessToken,
		httpusuario.AtualizarUsuarioRequest{Nome: "Bia S. Lima", Email: "bia.lima@oficina.com", Papel: string(shared.PapelAtendente)},
		&atualizado)
	if rec.Code != http.StatusOK {
		t.Fatalf("Atualizar: status %d, body %q", rec.Code, rec.Body.String())
	}
	if atualizado.Nome != "Bia S. Lima" || atualizado.Email != "bia.lima@oficina.com" || atualizado.Papel != string(shared.PapelAtendente) {
		t.Fatalf("Atualizar: dados não bateram: %+v", atualizado)
	}
	// email antigo não existe mais (não duplicou linha, só mudou o valor)
	rec = doRequest(t, http.MethodPost, "/v1/auth/login", "", httpauth.LoginRequest{Email: "bia@oficina.com", Senha: "senhaDefinitiva123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login com email antigo pós-Atualizar deveria ser 401, veio %d", rec.Code)
	}
	// login com o email novo continua funcionando com a mesma senha
	loginEmailNovo := doLogin(t, "bia.lima@oficina.com", "senhaDefinitiva123")
	if loginEmailNovo.AccessToken == "" {
		t.Fatal("login com email novo deveria funcionar")
	}

	// 5. Inativar: bloqueia login e derruba sessão em curso
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/usuarios/%d/inativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Inativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, http.MethodGet, "/v1/usuarios/me", loginEmailNovo.AccessToken, nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("Me pós-Inativar deveria ser 401, veio %d", rec.Code)
	}
	rec = doRequest(t, http.MethodPost, "/v1/auth/login", "", httpauth.LoginRequest{Email: "bia.lima@oficina.com", Senha: "senhaDefinitiva123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login pós-Inativar deveria ser 401, veio %d", rec.Code)
	}

	// 6. Ativar: reverte a inativação, login volta a funcionar
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/usuarios/%d/ativar", criado.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("Ativar: status %d, body %q", rec.Code, rec.Body.String())
	}
	loginPosReativacao := doLogin(t, "bia.lima@oficina.com", "senhaDefinitiva123")
	if loginPosReativacao.AccessToken == "" {
		t.Fatal("login pós-Ativar deveria funcionar de novo")
	}
	var mePosReativacao httpusuario.UsuarioResponse
	rec = doRequest(t, http.MethodGet, "/v1/usuarios/me", loginPosReativacao.AccessToken, nil, &mePosReativacao)
	if rec.Code != http.StatusOK {
		t.Fatalf("Me pós-Ativar: status %d", rec.Code)
	}
	if !mePosReativacao.Ativo || mePosReativacao.Nome != "Bia S. Lima" || mePosReativacao.Papel != string(shared.PapelAtendente) {
		t.Fatalf("Me pós-Ativar: dados da atualização não persistiram: %+v", mePosReativacao)
	}
}
