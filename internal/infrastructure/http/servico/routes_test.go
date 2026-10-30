package servico

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterServicoRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterServicoRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"POST /v1/servicos", "GET /v1/servicos", "GET /v1/servicos/:id",
		"PUT /v1/servicos/:id", "PATCH /v1/servicos/:id/ativar", "PATCH /v1/servicos/:id/inativar",
	}
	require.Len(t, routes, len(expected))
	for _, route := range expected {
		require.True(t, routes[route], route)
	}
}
