package auth_test

import (
	"testing"

	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/stretchr/testify/require"
)

func TestGerarRefreshTokenBruto_GeraValorNaoVazioEUnico(t *testing.T) {
	a, err := infraauth.GerarRefreshTokenBruto()
	require.NoError(t, err)
	require.NotEmpty(t, a)

	b, err := infraauth.GerarRefreshTokenBruto()
	require.NoError(t, err)
	require.NotEqual(t, a, b)
}

func TestHashRefreshToken_MesmaEntradaMesmoHash(t *testing.T) {
	bruto := "token-bruto-de-teste"

	require.Equal(t, infraauth.HashRefreshToken(bruto), infraauth.HashRefreshToken(bruto))
}

func TestHashRefreshToken_NuncaIgualAoBruto(t *testing.T) {
	bruto := "token-bruto-de-teste"

	require.NotEqual(t, bruto, infraauth.HashRefreshToken(bruto))
}
