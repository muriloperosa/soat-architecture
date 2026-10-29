//go:build integration

package integration_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"

	httpinfra "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/health"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// Rotas de health usadas pelas probes do Kubernetes (k8s/base/api-deployment.yaml):
// /v1/health/live na liveness e startup, /v1/health/ready na readiness.
// /v1/health continua existindo por compatibilidade.

func TestHealth_RotasRespondemOKComBancoDisponivel(t *testing.T) {
	for _, path := range []string{"/v1/health", "/v1/health/live", "/v1/health/ready"} {
		t.Run(path, func(t *testing.T) {
			var resp health.HealthCheckResponse
			rec := doRequest(t, http.MethodGet, path, "", nil, &resp)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, esperado 200 (body %q)", rec.Code, rec.Body.String())
			}
			if resp.Status != "ok" {
				t.Fatalf("status no corpo = %q, esperado \"ok\"", resp.Status)
			}
		})
	}
}

func TestHealth_RotasNaoExigemAutenticacao(t *testing.T) {
	// As probes do kubelet não mandam token. Se alguma dessas rotas passar a
	// exigir autenticação, todos os pods ficam não prontos ou são reiniciados.
	for _, path := range []string{"/v1/health/live", "/v1/health/ready"} {
		t.Run(path, func(t *testing.T) {
			rec := doRequest(t, http.MethodGet, path, "", nil, nil)

			if rec.Code == http.StatusUnauthorized || rec.Code == http.StatusForbidden {
				t.Fatalf("status = %d: rota de probe não pode exigir autenticação", rec.Code)
			}
		})
	}
}

// routerComBancoInacessivel monta o router de produção sobre uma conexão
// GORM que aponta para uma porta sem MySQL. Simula o banco fora do ar sem
// derrubar o container compartilhado pelos outros testes do pacote.
func routerComBancoInacessivel(t *testing.T) *gin.Engine {
	t.Helper()

	dsn := "root:root@tcp(127.0.0.1:1)/inexistente?timeout=1s"
	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		DSN:                       dsn,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	if err != nil {
		t.Fatalf("erro ao preparar conexão GORM: %v", err)
	}

	return httpinfra.NewRouter(wiring.NewContainer(testContainer.Config, db))
}

func TestHealthReady_Retorna503QuandoBancoInacessivel(t *testing.T) {
	router := routerComBancoInacessivel(t)

	for _, path := range []string{"/v1/health", "/v1/health/ready"} {
		t.Run(path, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, req)

			if rec.Code != http.StatusServiceUnavailable {
				t.Fatalf("status = %d, esperado 503 (body %q)", rec.Code, rec.Body.String())
			}

			var resp httperror.ErrorResponse
			if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
				t.Fatalf("erro ao decodificar resposta: %v", err)
			}
			if resp.Message != "Banco de dados relacional indisponível." {
				t.Fatalf("mensagem = %q", resp.Message)
			}
		})
	}
}

func TestHealthLive_Retorna200QuandoBancoInacessivel(t *testing.T) {
	// Garantia principal da liveness: banco fora do ar não pode fazer o
	// kubelet reiniciar a API.
	router := routerComBancoInacessivel(t)

	req := httptest.NewRequest(http.MethodGet, "/v1/health/live", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, esperado 200 (body %q)", rec.Code, rec.Body.String())
	}
}
