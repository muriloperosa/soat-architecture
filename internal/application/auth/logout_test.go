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

func TestLogoutUseCase_Executar_TokenExistente_Revoga(t *testing.T) {
	refreshTokens := &refreshTokensFake{}
	bruto, _ := infraauth.GerarRefreshTokenBruto()
	refreshTokens.salvos = append(refreshTokens.salvos, &domainauth.RefreshToken{
		ID: "rt-1", TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	})
	uc := appauth.NewLogoutUseCase(refreshTokens)

	err := uc.Executar(context.Background(), appauth.LogoutInput{RefreshTokenBruto: bruto})

	require.NoError(t, err)
	require.NotNil(t, refreshTokens.salvos[0].RevogadoEm)
}

func TestLogoutUseCase_Executar_TokenInexistente_NoOpSemErro(t *testing.T) {
	uc := appauth.NewLogoutUseCase(&refreshTokensFake{})

	err := uc.Executar(context.Background(), appauth.LogoutInput{RefreshTokenBruto: "inexistente"})

	require.NoError(t, err)
}
