package auth

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterAuthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterAuthRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	expected := []string{
		"POST /v1/auth/login",
		"POST /v1/auth/refresh",
		"POST /v1/auth/logout",
		"POST /v1/auth/cliente/login",
		"POST /v1/auth/cliente/refresh",
		"POST /v1/auth/cliente/logout",
	}
	require.Len(t, routes, len(expected))
	for _, route := range expected {
		require.True(t, routes[route], route)
	}
}
