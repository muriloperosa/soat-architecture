package cliente

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterClienteRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterClienteRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"POST /v1/clientes", "GET /v1/clientes", "GET /v1/clientes/documento/:documento",
		"GET /v1/clientes/:id", "PUT /v1/clientes/:id", "PATCH /v1/clientes/:id/ativar",
		"PATCH /v1/clientes/:id/inativar", "PUT /v1/clientes/me/senha",
	}
	require.Len(t, routes, len(expected))
	for _, route := range expected {
		require.True(t, routes[route], route)
	}
}
