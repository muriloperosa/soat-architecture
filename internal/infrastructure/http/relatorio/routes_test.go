package relatorio

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
)

func TestRegisterRelatorioRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	RegisterRelatorioRoutes(engine.Group("/v1"), &wiring.Container{})

	routes := engine.Routes()
	require.Len(t, routes, 1)
	require.Equal(t, "GET", routes[0].Method)
	require.Equal(t, "/v1/relatorios/ordens-servico/transicao-status", routes[0].Path)
}
