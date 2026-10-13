package ordemservico

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterOrdemServicoRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterOrdemServicoRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"GET /v1/ordens-servico", "GET /v1/ordens-servico/numero/:numero", "GET /v1/ordens-servico/:id",
		"POST /v1/ordens-servico", "PATCH /v1/ordens-servico/:id/entregar",
		"PATCH /v1/ordens-servico/:id/iniciar-diagnostico", "PUT /v1/ordens-servico/:id/diagnostico",
		"PATCH /v1/ordens-servico/:id/iniciar-execucao", "PATCH /v1/ordens-servico/:id/finalizar",
	}
	require.Len(t, routes, len(expected))
	for _, route := range expected {
		require.True(t, routes[route], route)
	}
}
