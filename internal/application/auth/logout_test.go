package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/auth/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestLogoutUseCase_Executar_TokenExistente_Revoga(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	bruto := "refresh-token-bruto-de-teste-1"
	rt := &domainauth.RefreshToken{
		ID: 1, TokenHash: domainauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	}
	uc := appauth.NewLogoutUseCase(refreshTokensRepo)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, domainauth.HashRefreshToken(bruto)).Return(rt, nil)
	refreshTokensRepo.EXPECT().Revogar(mock.Anything, uint64(1)).Return(nil)

	err := uc.Executar(context.Background(), appauth.LogoutInput{RefreshTokenBruto: bruto})

	require.NoError(t, err)
}

func TestLogoutUseCase_Executar_TokenJaRevogado_NoOpSemErro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	bruto := "refresh-token-bruto-de-teste-2"
	agora := time.Now()
	rt := &domainauth.RefreshToken{
		ID: 1, TokenHash: domainauth.HashRefreshToken(bruto),
		ExpiraEm: time.Now().Add(time.Hour), RevogadoEm: &agora,
	}
	uc := appauth.NewLogoutUseCase(refreshTokensRepo)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, domainauth.HashRefreshToken(bruto)).Return(rt, nil)

	err := uc.Executar(context.Background(), appauth.LogoutInput{RefreshTokenBruto: bruto})

	require.NoError(t, err)
}

func TestLogoutUseCase_Executar_TokenInexistente_NoOpSemErro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	uc := appauth.NewLogoutUseCase(refreshTokensRepo)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, domainauth.HashRefreshToken("inexistente")).Return(nil, errors.New("nao encontrado"))

	err := uc.Executar(context.Background(), appauth.LogoutInput{RefreshTokenBruto: "inexistente"})

	require.NoError(t, err)
}
