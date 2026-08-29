package auth_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/stretchr/testify/require"
)

func TestRefreshToken_EstaValido_TokenAtivoENaoExpirado(t *testing.T) {
	rt := auth.RefreshToken{ExpiraEm: time.Now().Add(1 * time.Hour)}

	require.True(t, rt.EstaValido())
}

func TestRefreshToken_EstaValido_TokenRevogado(t *testing.T) {
	revogadoEm := time.Now()
	rt := auth.RefreshToken{ExpiraEm: time.Now().Add(1 * time.Hour), RevogadoEm: &revogadoEm}

	require.False(t, rt.EstaValido())
}

func TestRefreshToken_EstaValido_TokenExpirado(t *testing.T) {
	rt := auth.RefreshToken{ExpiraEm: time.Now().Add(-1 * time.Hour)}

	require.False(t, rt.EstaValido())
}

func TestHashRefreshToken_MesmaEntradaMesmoHash(t *testing.T) {
	bruto := "token-bruto-de-teste"

	require.Equal(t, auth.HashRefreshToken(bruto), auth.HashRefreshToken(bruto))
}

func TestHashRefreshToken_NuncaIgualAoBruto(t *testing.T) {
	bruto := "token-bruto-de-teste"

	require.NotEqual(t, bruto, auth.HashRefreshToken(bruto))
}
