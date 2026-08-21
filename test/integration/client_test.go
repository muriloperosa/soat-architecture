//go:build integration

package integration_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	httpauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/auth"
)

// doRequest monta e executa um request HTTP contra o router real (com banco
// real por trás), decodifica a resposta em out se out != nil. body == nil
// significa sem corpo (GET, PATCH sem payload).
func doRequest(t *testing.T, method, path, bearer string, body any, out any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("erro ao serializar body: %v", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	if bearer != "" {
		req.Header.Set("Authorization", fmtBearer(bearer))
	}

	rec := httptest.NewRecorder()
	testRouter.ServeHTTP(rec, req)

	if out != nil && rec.Body.Len() > 0 {
		if err := json.Unmarshal(rec.Body.Bytes(), out); err != nil {
			t.Fatalf("erro ao decodificar resposta (status %d, body %q): %v", rec.Code, rec.Body.String(), err)
		}
	}
	return rec
}

// doLogin faz POST /v1/auth/login e falha o teste se não vier 200. Usado
// pelos testes que só precisam do token pra prosseguir, não estão testando
// o login em si.
func doLogin(t *testing.T, email, senha string) httpauth.TokenResponse {
	t.Helper()
	var resp httpauth.TokenResponse
	rec := doRequest(t, http.MethodPost, "/v1/auth/login", "", httpauth.LoginRequest{Email: email, Senha: senha}, &resp)
	if rec.Code != http.StatusOK {
		t.Fatalf("login falhou: status %d, body %q", rec.Code, rec.Body.String())
	}
	return resp
}

// doLoginCliente faz POST /v1/auth/cliente/login e falha o teste se não
// vier 200. Mantém separado o endpoint da fonte de identidade interna.
func doLoginCliente(t *testing.T, email, senha string) httpauth.TokenResponse {
	t.Helper()
	var resp httpauth.TokenResponse
	rec := doRequest(t, http.MethodPost, "/v1/auth/cliente/login", "", httpauth.LoginRequest{Email: email, Senha: senha}, &resp)
	if rec.Code != http.StatusOK {
		t.Fatalf("login de cliente falhou: status %d, body %q", rec.Code, rec.Body.String())
	}
	return resp
}

// Retorna bearer token formatado.
func fmtBearer(token string) string {
	return fmt.Sprintf("Bearer %s", token)
}
