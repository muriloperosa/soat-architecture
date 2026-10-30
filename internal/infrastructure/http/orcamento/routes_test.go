package orcamento

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterOrcamentoRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterOrcamentoRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"POST /v1/ordens-servico/:id/orcamento", "POST /v1/ordens-servico/:id/orcamento/itens-servico",
		"POST /v1/ordens-servico/:id/orcamento/itens-peca", "DELETE /v1/ordens-servico/:id/orcamento/itens-servico/:itemId",
		"DELETE /v1/ordens-servico/:id/orcamento/itens-peca/:itemId", "PATCH /v1/ordens-servico/:id/orcamento/itens-peca/:itemId/quantidade",
		"PATCH /v1/ordens-servico/:id/orcamento/enviar-aprovacao", "PATCH /v1/ordens-servico/:id/orcamento/aprovar",
		"PATCH /v1/ordens-servico/:id/orcamento/rejeitar",
	}
	require.Len(t, routes, len(expected))
	for _, route := range expected {
		require.True(t, routes[route], route)
	}
}
