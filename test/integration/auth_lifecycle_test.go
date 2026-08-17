//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
	httpusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/usuario"
)

// TestAuthLifecycle_LoginRefreshLogout prova a cadeia real
// AuthenticationMiddleware + RefreshTokenRepository + CredenciaisAdapter +
// GORM ao longo de um ciclo completo.
func TestAuthLifecycle_LoginRefreshLogout(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)

	// login: credenciais válidas emitem o par de tokens
	login := doLogin(t, "ana@oficina.com", "senha123")
	if login.AccessToken == "" || login.RefreshToken == "" {
		t.Fatalf("login não retornou tokens: %+v", login)
	}

	// access token do login acessa rota protegida
	var me httpusuario.UsuarioResponse
	rec := doRequest(t, http.MethodGet, "/v1/usuarios/me", login.AccessToken, nil, &me)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /me com token do login: status %d, body %q", rec.Code, rec.Body.String())
	}
	if me.Email != "ana@oficina.com" {
		t.Fatalf("GET /me retornou usuário errado: %+v", me)
	}

	// refresh: token antigo rotaciona pra um novo par
	var refreshed httpauth.TokenResponse
	rec = doRequest(t, http.MethodPost, "/v1/auth/refresh", "", httpauth.RefreshRequest{RefreshToken: login.RefreshToken}, &refreshed)
	if rec.Code != http.StatusOK {
		t.Fatalf("refresh falhou: status %d, body %q", rec.Code, rec.Body.String())
	}
	if refreshed.AccessToken == login.AccessToken {
		t.Fatal("refresh deveria emitir um access token novo, veio o mesmo")
	}

	// access token novo (pós-refresh) continua acessando rota protegida
	rec = doRequest(t, http.MethodGet, "/v1/usuarios/me", refreshed.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /me com token pós-refresh: status %d, body %q", rec.Code, rec.Body.String())
	}

	// logout: revoga o refresh token novo
	rec = doRequest(t, http.MethodPost, "/v1/auth/logout", "", httpauth.RefreshRequest{RefreshToken: refreshed.RefreshToken}, nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("logout falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	// access token pós-logout para de funcionar: AccessTokenRevogado pareia
	// pelo jti do refresh token revogado, não precisa esperar o access
	// token expirar sozinho
	rec = doRequest(t, http.MethodGet, "/v1/usuarios/me", refreshed.AccessToken, nil, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET /me pós-logout deveria ser 401, veio %d", rec.Code)
	}
}
