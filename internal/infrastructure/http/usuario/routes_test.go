package usuario

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterUsuarioRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterUsuarioRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"POST /v1/usuarios", "PUT /v1/usuarios/:id", "PATCH /v1/usuarios/:id/ativar",
		"PATCH /v1/usuarios/:id/inativar", "GET /v1/usuarios/me", "PUT /v1/usuarios/me/senha",
	}
	require.Len(t, routes, len(expected))
	for _, route := range expected {
		require.True(t, routes[route], route)
	}
}
