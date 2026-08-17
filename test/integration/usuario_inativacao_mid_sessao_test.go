//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
)

// TestInativacaoMidSessao_TokenAntigoParaDeFuncionar é o teste que
// realmente prova o EstaAtivo por request (AuthenticationMiddleware +
// UsuarioStatusRepository): um usuário loga, guarda o access token, um
// admin inativa esse usuário, e a mesma requisição com o token antigo tem
// que virar 401 sem esperar o token expirar.
func TestInativacaoMidSessao_TokenAntigoParaDeFuncionar(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	alvo := seedUsuario(t, "Bia Lima", "bia@oficina.com", "senha123", shared.PapelMecanico)

	// alvo loga e guarda o access token, este token não vai ser renovado
	// de novo, é o mesmo usado depois da inativação
	loginAlvo := doLogin(t, "bia@oficina.com", "senha123")

	// confirma que o token funciona antes da inativação
	rec := doRequest(t, http.MethodGet, "/v1/usuarios/me", loginAlvo.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /me antes de inativar: status %d, esperava 200", rec.Code)
	}

	// admin loga e inativa o alvo
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	rec = doRequest(t, http.MethodPatch, fmt.Sprintf("/v1/usuarios/%d/inativar", alvo.ID), loginAdmin.AccessToken, nil, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("PATCH /inativar falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	// mesmo token, emitido antes da inativação, agora deve ser rejeitado
	// isso é o que garante que inativar derruba sessão em curso, não só
	// bloqueia login novo
	rec = doRequest(t, http.MethodGet, "/v1/usuarios/me", loginAlvo.AccessToken, nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /me pós-inativação deveria ser 401, veio %d, body %q", rec.Code, rec.Body.String())
	}

	// e login novo também é bloqueado (Credencial.Ativo=false no LoginUseCase)
	rec = doRequest(t, http.MethodPost, "/v1/auth/login", "", httpauth.LoginRequest{Email: "bia@oficina.com", Senha: "senha123"}, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("login pós-inativação deveria ser 401, veio %d", rec.Code)
	}
}
