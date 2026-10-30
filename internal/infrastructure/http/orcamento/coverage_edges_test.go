package orcamento_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	"github.com/stretchr/testify/require"
)

func TestHandlersOrcamento_RejeitamIDsInvalidos(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := httporcamento.NewHandler(nil, nil, nil, nil, nil, nil, nil)
	tests := []struct {
		nome, method, route, path string
		call                      gin.HandlerFunc
	}{
		{"gerar", http.MethodPost, "/:id", "/invalid", h.Gerar},
		{"adicionar-servico", http.MethodPost, "/:id/servicos", "/invalid/servicos", h.AdicionarServico},
		{"adicionar-peca", http.MethodPost, "/:id/pecas", "/invalid/pecas", h.AdicionarPeca},
		{"alterar-quantidade", http.MethodPatch, "/:id/pecas/:itemId", "/invalid/pecas/1", h.AlterarQuantidadePeca},
	}
	for _, tt := range tests {
		t.Run(tt.nome, func(t *testing.T) {
			router := gin.New()
			router.Handle(tt.method, tt.route, tt.call)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.path, nil))
			require.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestDecisaoHandler_RejeitarRejeitaIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := httporcamento.NewDecisaoHandler(nil, nil)
	router := gin.New()
	router.PATCH("/:id/rejeitar", h.Rejeitar)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/invalid/rejeitar", nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerOrcamento_AlterarQuantidadeRejeitaItemInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := httporcamento.NewHandler(nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.PATCH("/:id/pecas/:itemId", h.AlterarQuantidadePeca)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/1/pecas/invalido", nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
