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
	"golang.org/x/crypto/bcrypt"
)

func hashSenha(t *testing.T, senha string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(h)
}

func TestLoginUseCase_Executar_CredencialValida_RetornaTokens(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoCliente, 168*time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: "user-1", SenhaHash: hashSenha(t, "senha123"), Papel: domainauth.PapelCliente}, nil)

	var rtSalvo *domainauth.RefreshToken
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Run(func(ctx context.Context, rt *domainauth.RefreshToken) {
			rt.ID = "rt-1"
			rtSalvo = rt
		}).
		Return(nil)

	out, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.NoError(t, err)
	require.NotEmpty(t, out.AccessToken)
	require.NotEmpty(t, out.RefreshToken)
	require.NotNil(t, rtSalvo)
	require.Equal(t, "user-1", rtSalvo.UsuarioID)
	require.Equal(t, domainauth.TipoCliente, rtSalvo.Tipo)
	require.Equal(t, domainauth.PapelCliente, rtSalvo.Papel)

	claims, err := jwtAuth.ValidarAccessToken(out.AccessToken)
	require.NoError(t, err)
	require.Equal(t, domainauth.PapelCliente, claims.Papel)
}

func TestLoginUseCase_Executar_UsuarioInterno_PropagaPapel(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoInterno, 168*time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "m@a.com").
		Return(&domainauth.Credencial{ID: "user-2", SenhaHash: hashSenha(t, "senha123"), Papel: domainauth.PapelMecanico}, nil)

	var rtSalvo *domainauth.RefreshToken
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Run(func(ctx context.Context, rt *domainauth.RefreshToken) {
			rt.ID = "rt-2"
			rtSalvo = rt
		}).
		Return(nil)

	out, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "m@a.com", Senha: "senha123"})

	require.NoError(t, err)
	claims, err := jwtAuth.ValidarAccessToken(out.AccessToken)
	require.NoError(t, err)
	require.Equal(t, domainauth.PapelMecanico, claims.Papel)
	require.NotNil(t, rtSalvo)
	require.Equal(t, domainauth.PapelMecanico, rtSalvo.Papel)
}

func TestLoginUseCase_Executar_SenhaErrada_ErroGenerico(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, infraauth.NewAuthenticatorJWT("s", time.Minute), domainauth.TipoCliente, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: "user-1", SenhaHash: hashSenha(t, "senha123")}, nil)

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "errada"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credenciais inválidas")
}

func TestLoginUseCase_Executar_ErroAoGerarAccessToken_RetornaErroInterno(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoCliente, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: "user-1", SenhaHash: hashSenha(t, "senha123"), Papel: domainauth.PapelCliente}, nil)
	jwtAuth.EXPECT().
		GerarAccessToken("user-1", domainauth.TipoCliente, domainauth.PapelCliente).
		Return("", errors.New("chave de assinatura invalida"))

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao gerar access token")
}

func TestLoginUseCase_Executar_ErroAoGerarRefreshTokenBruto_RetornaErroInterno(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoCliente, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: "user-1", SenhaHash: hashSenha(t, "senha123"), Papel: domainauth.PapelCliente}, nil)
	jwtAuth.EXPECT().
		GerarAccessToken("user-1", domainauth.TipoCliente, domainauth.PapelCliente).
		Return("access-token-valido", nil)
	jwtAuth.EXPECT().
		GerarRefreshTokenBruto().
		Return("", errors.New("entropia insuficiente"))

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao gerar refresh token")
}

func TestLoginUseCase_Executar_ErroAoSalvarRefreshToken_RetornaErroInterno(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, infraauth.NewAuthenticatorJWT("s", time.Minute), domainauth.TipoCliente, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: "user-1", SenhaHash: hashSenha(t, "senha123"), Papel: domainauth.PapelCliente}, nil)
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao salvar refresh token")
}

func TestLoginUseCase_Executar_EmailNaoEncontrado_MesmoErroGenerico(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, infraauth.NewAuthenticatorJWT("s", time.Minute), domainauth.TipoCliente, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "naoexiste@a.com").
		Return(nil, errors.New("nao encontrado"))

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "naoexiste@a.com", Senha: "x"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credenciais inválidas")
}
