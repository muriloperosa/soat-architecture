package orcamento

import (
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
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

func TestRegisterOrcamentoRoutes_RegistraRotasPublicasQuandoTokenConfigurado(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	provider, err := infraauth.NewApprovalTokenJWT("segredo-de-teste", 24*time.Hour)
	require.NoError(t, err)
	container := &wiring.Container{
		Config:        &config.Config{PublicBaseURL: "https://oficina.example"},
		ApprovalToken: provider,
	}
	RegisterOrcamentoRoutes(engine.Group("/v1"), container)

	routes := make(map[string]bool)
	for _, route := range engine.Routes() {
		routes[route.Method+" "+route.Path] = true
	}
	require.True(t, routes["GET /v1/orcamentos/aprovacao"])
	require.True(t, routes["POST /v1/orcamentos/aprovacao/decisao"])
}
