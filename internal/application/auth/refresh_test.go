package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/auth/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func gerarBrutoDeTeste(t *testing.T) string {
	t.Helper()
	bruto, err := infraauth.NewAuthenticatorJWT("s", time.Minute).GerarRefreshToken()
	require.NoError(t, err)
	return bruto
}

func TestRefreshUseCase_Executar_TokenValido_RotacionaERetornaNovoPar(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	bruto := gerarBrutoDeTeste(t)
	rtAntigo := &domainauth.RefreshToken{
		ID: 1, UsuarioID: 1, Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente,
		TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	}
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, jwtAuth, time.Hour, time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rtAntigo, nil)
	refreshTokensRepo.EXPECT().Revogar(mock.Anything, uint64(1)).Return(nil)
	jwtAuth.EXPECT().GerarAccessToken("1", domainauth.TipoCliente, shared.PapelCliente).Return("novo-access-token", "jti-novo", nil)
	jwtAuth.EXPECT().GerarRefreshToken().Return("novo-refresh-bruto", nil)
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Run(func(ctx context.Context, rt *domainauth.RefreshToken) { rt.ID = 2 }).
		Return(nil)

	out, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.NoError(t, err)
	require.Equal(t, "novo-access-token", out.AccessToken)
	require.Equal(t, "novo-refresh-bruto", out.RefreshToken)
}

func TestRefreshUseCase_Executar_ErroAoRevogar_RetornaErroInterno(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	bruto := gerarBrutoDeTeste(t)
	rt := &domainauth.RefreshToken{
		ID: 1, TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	}
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, jwtAuth, time.Hour, time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rt, nil)
	refreshTokensRepo.EXPECT().Revogar(mock.Anything, uint64(1)).Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao revogar refresh token")
}

func TestRefreshUseCase_Executar_ErroAoSalvarNovoPar_PropagaErro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	bruto := gerarBrutoDeTeste(t)
	rt := &domainauth.RefreshToken{
		ID: 1, UsuarioID: 1, Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente,
		TokenHash: infraauth.HashRefreshToken(bruto), ExpiraEm: time.Now().Add(time.Hour),
	}
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, jwtAuth, time.Hour, time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rt, nil)
	refreshTokensRepo.EXPECT().Revogar(mock.Anything, uint64(1)).Return(nil)
	jwtAuth.EXPECT().GerarAccessToken("1", domainauth.TipoCliente, shared.PapelCliente).Return("novo-access-token", "jti-novo", nil)
	jwtAuth.EXPECT().GerarRefreshToken().Return("novo-refresh-bruto", nil)
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao salvar refresh token")
}

func TestRefreshUseCase_Executar_TokenJaRevogado_Erro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	bruto := gerarBrutoDeTeste(t)
	agora := time.Now()
	rt := &domainauth.RefreshToken{
		ID: 1, TokenHash: infraauth.HashRefreshToken(bruto),
		ExpiraEm: time.Now().Add(time.Hour), RevogadoEm: &agora,
	}
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, jwtAuth, time.Hour, time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken(bruto)).Return(rt, nil)

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: bruto})

	require.Error(t, err)
}

func TestRefreshUseCase_Executar_TokenInexistente_Erro(t *testing.T) {
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	uc := appauth.NewRefreshUseCase(refreshTokensRepo, jwtAuth, time.Hour, time.Hour)

	refreshTokensRepo.EXPECT().BuscarPorHash(mock.Anything, infraauth.HashRefreshToken("inexistente")).Return(nil, errors.New("nao encontrado"))

	_, err := uc.Executar(context.Background(), appauth.RefreshInput{RefreshTokenBruto: "inexistente"})

	require.Error(t, err)
}
