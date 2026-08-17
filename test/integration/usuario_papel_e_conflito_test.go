//go:build integration

package integration_test

import (
	"net/http"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/usuario"
)

// TestAuthorization_NaoAdminNaoAcessaRotaAdmin prova o AuthorizationMiddleware
// variádico (checagem de papel) contra o banco de verdade: um mecânico
// autenticado, tentando uma rota restrita a admin, tem que levar 403, não
// 401 (ele está autenticado, só não tem o papel certo) nem 200.
func TestAuthorization_NaoAdminNaoAcessaRotaAdmin(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Bia Lima", "bia@oficina.com", "senha123", shared.PapelMecanico)
	login := doLogin(t, "bia@oficina.com", "senha123")

	rec := doRequest(t, http.MethodPost, "/v1/usuarios", login.AccessToken,
		httpusuario.CriarUsuarioRequest{Nome: "Outro", Email: "outro@oficina.com", Senha: "senha123", Papel: shared.PapelAtendente},
		nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST /v1/usuarios por não-admin deveria ser 403, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

// TestCriarUsuario_EmailDuplicado_Retorna409 prova a checagem de unicidade
// de email fim a fim: o segundo POST com o mesmo email tem que ser rejeitado
// com 409 (não 500 de constraint do banco vazando) e não pode deixar linha
// órfã nem duplicata na tabela.
func TestCriarUsuario_EmailDuplicado_Retorna409(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	body := httpusuario.CriarUsuarioRequest{Nome: "Bia Lima", Email: "bia@oficina.com", Senha: "senha123", Papel: shared.PapelMecanico}

	rec := doRequest(t, http.MethodPost, "/v1/usuarios", login.AccessToken, body, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("primeira criação falhou: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPost, "/v1/usuarios", login.AccessToken, body, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("email duplicado deveria ser 409, veio %d, body %q", rec.Code, rec.Body.String())
	}

	var count int64
	if err := testDB.Table("usuarios").Where("email = ?", "bia@oficina.com").Count(&count).Error; err != nil {
		t.Fatalf("erro ao contar usuarios: %v", err)
	}
	if count != 1 {
		t.Fatalf("esperava exatamente 1 linha com esse email, achou %d", count)
	}
}
