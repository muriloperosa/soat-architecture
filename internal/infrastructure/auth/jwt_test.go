package auth_test

import (
	"testing"
	"time"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/stretchr/testify/require"
)

func TestGerarEValidarAccessToken_TokenValido(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	token, err := autenticador.GerarAccessToken("user-123", domainauth.TipoCliente, domainauth.PapelCliente)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := autenticador.ValidarAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, "user-123", claims.Subject)
	require.Equal(t, domainauth.TipoCliente, claims.Tipo)
	require.Equal(t, domainauth.PapelCliente, claims.Papel)
}

func TestValidarAccessToken_TokenExpirado(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", -1*time.Minute)

	token, err := autenticador.GerarAccessToken("user-123", domainauth.TipoInterno, domainauth.PapelAdmin)
	require.NoError(t, err)

	_, err = autenticador.ValidarAccessToken(token)
	require.Error(t, err)
}

func TestValidarAccessToken_AssinaturaInvalida(t *testing.T) {
	gerador := infraauth.NewAuthenticatorJWT("segredo-a", 15*time.Minute)
	validador := infraauth.NewAuthenticatorJWT("segredo-b", 15*time.Minute)

	token, err := gerador.GerarAccessToken("user-123", domainauth.TipoInterno, domainauth.PapelAdmin)
	require.NoError(t, err)

	_, err = validador.ValidarAccessToken(token)
	require.Error(t, err)
}
