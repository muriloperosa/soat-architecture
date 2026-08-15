package auth_test

import (
	"context"
	"testing"
	"time"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/stretchr/testify/require"
)

func TestRefreshUseCase_Executar_TokenValido_RotacionaERetornaNovoPar(t *testing.T) {
	refreshTokens := &refreshTokensFake{}
	bruto, _ := infraauth.GerarRefreshTokenBruto()
	refreshTokens.salvos = append(refreshTokens.salvos, &domainauth.RefreshToken{
		ID: "rt-1", UsuarioID: "user-1", Tipo: domainauth.TipoCliente,
		TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	})
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	uc := appauth.NewRefreshUseCase(refreshTokens, jwtAuth, time.Hour)

	out, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.NoError(t, err)
	require.NotEmpty(t, out.AccessToken)
	require.NotEqual(t, bruto, out.RefreshToken)
	require.NotNil(t, refreshTokens.salvos[0].RevogadoEm, "token antigo deve ser revogado")
	require.Len(t, refreshTokens.salvos, 2, "novo par deve ser persistido")
}

func TestRefreshUseCase_Executar_TokenJaRevogado_Erro(t *testing.T) {
	refreshTokens := &refreshTokensFake{}
	bruto, _ := infraauth.GerarRefreshTokenBruto()
	agora := time.Now()
	refreshTokens.salvos = append(refreshTokens.salvos, &domainauth.RefreshToken{
		ID: "rt-1", TokenHash: infraauth.HashRefreshToken(bruto),
		ExpiraEm: time.Now().Add(time.Hour), RevogadoEm: &agora,
	})
	uc := appauth.NewRefreshUseCase(refreshTokens, infraauth.NewAuthenticatorJWT("s", time.Minute), time.Hour)

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.Error(t, err)
}

func TestRefreshUseCase_Executar_TokenInexistente_Erro(t *testing.T) {
	uc := appauth.NewRefreshUseCase(&refreshTokensFake{}, infraauth.NewAuthenticatorJWT("s", time.Minute), time.Hour)

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: "inexistente"})

	require.Error(t, err)
}
