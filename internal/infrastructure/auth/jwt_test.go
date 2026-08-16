package auth_test

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/stretchr/testify/require"
)

func TestGerarEValidarAccessToken_TokenValido(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	token, jti, err := autenticador.GerarAccessToken("user-123", domainauth.TipoCliente, shared.PapelCliente)
	require.NoError(t, err)
	require.NotEmpty(t, token)
	require.NotEmpty(t, jti)

	claims, err := autenticador.ValidarAccessToken(token)
	require.NoError(t, err)
	require.Equal(t, "user-123", claims.Subject)
	require.Equal(t, domainauth.TipoCliente, claims.Tipo)
	require.Equal(t, shared.PapelCliente, claims.Papel)
	require.Equal(t, jti, claims.Jti)
}

func TestGerarAccessToken_GeraJtiDiferenteACadaChamada(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	_, jtiA, err := autenticador.GerarAccessToken("user-123", domainauth.TipoCliente, shared.PapelCliente)
	require.NoError(t, err)

	_, jtiB, err := autenticador.GerarAccessToken("user-123", domainauth.TipoCliente, shared.PapelCliente)
	require.NoError(t, err)

	require.NotEqual(t, jtiA, jtiB)
}

func TestValidarAccessToken_TokenExpirado(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", -1*time.Minute)

	token, _, err := autenticador.GerarAccessToken("user-123", domainauth.TipoInterno, shared.PapelAdmin)
	require.NoError(t, err)

	_, err = autenticador.ValidarAccessToken(token)
	require.Error(t, err)
}

func TestValidarAccessToken_AssinaturaInvalida(t *testing.T) {
	gerador := infraauth.NewAuthenticatorJWT("segredo-a", 15*time.Minute)
	validador := infraauth.NewAuthenticatorJWT("segredo-b", 15*time.Minute)

	token, _, err := gerador.GerarAccessToken("user-123", domainauth.TipoInterno, shared.PapelAdmin)
	require.NoError(t, err)

	_, err = validador.ValidarAccessToken(token)
	require.Error(t, err)
}

func TestValidarAccessToken_TokenMalformado(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	_, err := autenticador.ValidarAccessToken("isso-nao-e-um-jwt")
	require.Error(t, err)
}

func TestValidarAccessToken_TokenVazio(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	_, err := autenticador.ValidarAccessToken("")
	require.Error(t, err)
}

func TestValidarAccessToken_MetodoDeAssinaturaNaoHMAC_Rejeitado(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	// Forja um token com alg "none" (sem assinatura) — se o middleware não
	// checasse o tipo do método, um atacante poderia forjar claims livremente.
	claims := domainauth.AppClaims{
		Subject: "user-123",
		Tipo:    domainauth.TipoInterno,
		Papel:   shared.PapelAdmin,
		Jti:     "jti-forjado",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "user-123",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	tokenForjado := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenBruto, err := tokenForjado.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = autenticador.ValidarAccessToken(tokenBruto)
	require.Error(t, err)
}

func TestGerarRefreshToken_GeraValorNaoVazioEUnico(t *testing.T) {
	autenticador := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	a, err := autenticador.GerarRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, a)

	b, err := autenticador.GerarRefreshToken()
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
