package auth

import (
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/require"
)

const approvalTestSecret = "segredo-proprio-de-aprovacao-para-testes"

func TestApprovalTokenJWT_GerarEValidarToken(t *testing.T) {
	provider, err := NewApprovalTokenJWT(approvalTestSecret, 24*time.Hour)
	require.NoError(t, err)

	token, err := provider.GerarToken(123, 456)
	require.NoError(t, err)
	require.NotEmpty(t, token)

	claims, err := provider.ValidarToken(token)
	require.NoError(t, err)
	require.Equal(t, uint64(123), claims.OrdemServicoID)
	require.Equal(t, uint64(456), claims.ClienteID)
	require.WithinDuration(t, time.Now().Add(24*time.Hour), claims.ExpiresAt, time.Minute)
}

func TestApprovalTokenJWT_RejeitaTokenExpirado(t *testing.T) {
	provider, err := NewApprovalTokenJWT(approvalTestSecret, -time.Minute)
	require.Error(t, err)
	require.Nil(t, provider)

	claims := approvalTokenClaims{
		OrdemServicoID: 123,
		ClienteID:      456,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-time.Minute)),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(approvalTestSecret))
	require.NoError(t, err)

	provider, err = NewApprovalTokenJWT(approvalTestSecret, 24*time.Hour)
	require.NoError(t, err)
	_, err = provider.ValidarToken(token)
	require.Error(t, err)
}

func TestApprovalTokenJWT_RejeitaAssinaturaDeOutroSegredo(t *testing.T) {
	issuer, err := NewApprovalTokenJWT(approvalTestSecret, 24*time.Hour)
	require.NoError(t, err)
	validator, err := NewApprovalTokenJWT("outro-segredo-proprio-de-aprovacao", 24*time.Hour)
	require.NoError(t, err)

	token, err := issuer.GerarToken(123, 456)
	require.NoError(t, err)
	_, err = validator.ValidarToken(token)
	require.Error(t, err)
}

func TestApprovalTokenJWT_RejeitaTokenMalformadoOuSemExpiracao(t *testing.T) {
	provider, err := NewApprovalTokenJWT(approvalTestSecret, 24*time.Hour)
	require.NoError(t, err)

	_, err = provider.ValidarToken("nao-e-um-token")
	require.Error(t, err)

	withoutExpiry := approvalTokenClaims{OrdemServicoID: 123, ClienteID: 456}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, withoutExpiry).SignedString([]byte(approvalTestSecret))
	require.NoError(t, err)
	_, err = provider.ValidarToken(token)
	require.Error(t, err)
}

func TestApprovalTokenJWT_RejeitaIdentificadoresZero(t *testing.T) {
	provider, err := NewApprovalTokenJWT(approvalTestSecret, 24*time.Hour)
	require.NoError(t, err)

	_, err = provider.GerarToken(0, 456)
	require.Error(t, err)
	_, err = provider.GerarToken(123, 0)
	require.Error(t, err)
}

func TestNewApprovalTokenJWT_RejeitaConfiguracaoInvalida(t *testing.T) {
	_, err := NewApprovalTokenJWT("", 24*time.Hour)
	require.Error(t, err)
	_, err = NewApprovalTokenJWT(approvalTestSecret, 0)
	require.Error(t, err)
}
