package health

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterHealthRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterHealthRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := engine.Routes()
	require.Len(t, routes, 3)
	paths := make([]string, 0, len(routes))
	for _, route := range routes {
		require.Equal(t, "GET", route.Method)
		paths = append(paths, route.Path)
	}
	require.ElementsMatch(t, []string{
		"/v1/health",
		"/v1/health/live",
		"/v1/health/ready",
	}, paths)
}
