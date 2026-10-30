package veiculo

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterVeiculoRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterVeiculoRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"POST /v1/veiculos", "PUT /v1/veiculos/:id", "PATCH /v1/veiculos/:id/ativar",
		"PATCH /v1/veiculos/:id/inativar", "GET /v1/veiculos/placa/:placa",
		"GET /v1/veiculos/:id", "GET /v1/veiculos",
	}
	require.Len(t, routes, len(expected))
	for _, route := range expected {
		require.True(t, routes[route], route)
	}
}
