package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/auth/mocks"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRefreshUseCase_Executar_TokenValido_RotacionaERetornaNovoPar(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	bruto, _ := infraauth.GerarRefreshTokenBruto()
	rtAntigo := &domainauth.RefreshToken{
		ID: "rt-1", UsuarioID: "user-1", Tipo: domainauth.TipoCliente,
		TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	}
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo", 15*time.Minute)
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, jwtAuth, time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rtAntigo, nil)
	refreshTokensRepo.EXPECT().Revogar(mock.Anything, "rt-1").Return(nil)
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Run(func(ctx context.Context, rt *domainauth.RefreshToken) { rt.ID = "rt-2" }).
		Return(nil)

	out, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.NoError(t, err)
	require.NotEmpty(t, out.AccessToken)
	require.NotEqual(t, bruto, out.RefreshToken)
}

func TestRefreshUseCase_Executar_ErroAoRevogar_RetornaErroInterno(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	bruto, _ := infraauth.GerarRefreshTokenBruto()
	rt := &domainauth.RefreshToken{
		ID: "rt-1", TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	}
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, infraauth.NewAuthenticatorJWT("s", time.Minute), time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rt, nil)
	refreshTokensRepo.EXPECT().Revogar(mock.Anything, "rt-1").Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao revogar refresh token")
}

func TestRefreshUseCase_Executar_ErroAoSalvarNovoPar_PropagaErro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	bruto, _ := infraauth.GerarRefreshTokenBruto()
	rt := &domainauth.RefreshToken{
		ID: "rt-1", TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	}
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, infraauth.NewAuthenticatorJWT("s", time.Minute), time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rt, nil)
	refreshTokensRepo.EXPECT().Revogar(mock.Anything, "rt-1").Return(nil)
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao salvar refresh token")
}

func TestRefreshUseCase_Executar_TokenJaRevogado_Erro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	bruto, _ := infraauth.GerarRefreshTokenBruto()
	agora := time.Now()
	rt := &domainauth.RefreshToken{
		ID: "rt-1", TokenHash: infraauth.HashRefreshToken(bruto),
		ExpiraEm: time.Now().Add(time.Hour), RevogadoEm: &agora,
	}
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, infraauth.NewAuthenticatorJWT("s", time.Minute), time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rt, nil)

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.Error(t, err)
}

func TestRefreshUseCase_Executar_TokenInexistente_Erro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, infraauth.NewAuthenticatorJWT("s", time.Minute), time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken("inexistente")).Return(nil, errors.New("nao encontrado"))

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: "inexistente"})

	require.Error(t, err)
}
