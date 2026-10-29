package auth

import (
	"errors"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestGerarEValidarAccessToken_TokenValido(t *testing.T) {
	autenticador := NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

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
	autenticador := NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	_, jtiA, err := autenticador.GerarAccessToken("user-123", domainauth.TipoCliente, shared.PapelCliente)
	require.NoError(t, err)

	_, jtiB, err := autenticador.GerarAccessToken("user-123", domainauth.TipoCliente, shared.PapelCliente)
	require.NoError(t, err)

	require.NotEqual(t, jtiA, jtiB)
}

func TestValidarAccessToken_TokenExpirado(t *testing.T) {
	autenticador := NewAuthenticatorJWT("segredo-de-teste", -1*time.Minute)

	token, _, err := autenticador.GerarAccessToken("user-123", domainauth.TipoInterno, shared.PapelAdmin)
	require.NoError(t, err)

	_, err = autenticador.ValidarAccessToken(token)
	require.Error(t, err)
}

func TestValidarAccessToken_AssinaturaInvalida(t *testing.T) {
	gerador := NewAuthenticatorJWT("segredo-a", 15*time.Minute)
	validador := NewAuthenticatorJWT("segredo-b", 15*time.Minute)

	token, _, err := gerador.GerarAccessToken("user-123", domainauth.TipoInterno, shared.PapelAdmin)
	require.NoError(t, err)

	_, err = validador.ValidarAccessToken(token)
	require.Error(t, err)
}

func TestValidarAccessToken_TokenMalformado(t *testing.T) {
	autenticador := NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	_, err := autenticador.ValidarAccessToken("isso-nao-e-um-jwt")
	require.Error(t, err)
}

func TestValidarAccessToken_TokenVazio(t *testing.T) {
	autenticador := NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	_, err := autenticador.ValidarAccessToken("")
	require.Error(t, err)
}

func TestValidarAccessToken_MetodoDeAssinaturaNaoHMAC_Rejeitado(t *testing.T) {
	autenticador := NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	// Forja um token com alg "none" (sem assinatura) — se o middleware não
	// checasse o tipo do método, um atacante poderia forjar claims livremente.
	claims := jwt.MapClaims{
		"sub":   "user-123",
		"tipo":  domainauth.TipoInterno,
		"papel": shared.PapelAdmin,
		"jti":   "jti-forjado",
		"exp":   jwt.NewNumericDate(time.Now().Add(time.Hour)),
	}
	tokenForjado := jwt.NewWithClaims(jwt.SigningMethodNone, claims)
	tokenBruto, err := tokenForjado.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	_, err = autenticador.ValidarAccessToken(tokenBruto)
	require.Error(t, err)
}

func TestGerarRefreshToken_GeraValorNaoVazioEUnico(t *testing.T) {
	autenticador := NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	a, err := autenticador.GerarRefreshToken()
	require.NoError(t, err)
	require.NotEmpty(t, a)

	b, err := autenticador.GerarRefreshToken()
	require.NoError(t, err)
	require.NotEqual(t, a, b)
}

func TestGerarAccessToken_PayloadTemOsClaimsEsperados(t *testing.T) {
	autenticador := NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)

	token, jti, err := autenticador.GerarAccessToken("user-123", domainauth.TipoInterno, shared.PapelMecanico)
	require.NoError(t, err)

	claims := jwt.MapClaims{}
	_, _, err = jwt.NewParser().ParseUnverified(token, claims)
	require.NoError(t, err)

	require.Equal(t, "user-123", claims["sub"])
	require.Equal(t, string(domainauth.TipoInterno), claims["tipo"])
	require.Equal(t, string(shared.PapelMecanico), claims["papel"])
	require.Equal(t, jti, claims["jti"])
	require.Contains(t, claims, "iat")
	require.Contains(t, claims, "exp")
}

func TestGerarRefreshTokenFalhaQuandoFonteAleatoriaFalha(t *testing.T) {
	autenticador := &AuthenticatorJWT{random: coverageFailingRandomReader{}}
	_, err := autenticador.GerarRefreshToken()
	require.Error(t, err)
}

type coverageFailingRandomReader struct{}

func (coverageFailingRandomReader) Read([]byte) (int, error) {
	return 0, errors.New("fonte aleatória indisponível")
}
