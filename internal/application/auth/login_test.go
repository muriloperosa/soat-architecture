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
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoCliente, 168*time.Hour, 168*time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: 1, SenhaHash: hashSenha(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}, nil)

	var rtSalvo *domainauth.RefreshToken
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Run(func(ctx context.Context, rt *domainauth.RefreshToken) {
			rt.ID = 101
			rtSalvo = rt
		}).
		Return(nil)

	out, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.NoError(t, err)
	require.NotEmpty(t, out.AccessToken)
	require.NotEmpty(t, out.RefreshToken)
	require.NotNil(t, rtSalvo)
	require.Equal(t, uint64(1), rtSalvo.UsuarioID)
	require.Equal(t, domainauth.TipoCliente, rtSalvo.Tipo)
	require.Equal(t, shared.PapelCliente, rtSalvo.Papel)
	require.NotEmpty(t, rtSalvo.AccessTokenJti)

	claims, err := jwtAuth.ValidarAccessToken(out.AccessToken)
	require.NoError(t, err)
	require.Equal(t, shared.PapelCliente, claims.Papel)
	require.Equal(t, rtSalvo.AccessTokenJti, claims.Jti)
}

func TestLoginUseCase_Executar_UsuarioInterno_PropagaPapel(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoInterno, 168*time.Hour, 168*time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "m@a.com").
		Return(&domainauth.Credencial{ID: 2, SenhaHash: hashSenha(t, "senha123"), Papel: shared.PapelMecanico, Ativo: true}, nil)

	var rtSalvo *domainauth.RefreshToken
	refreshTokensRepo.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*auth.RefreshToken")).
		Run(func(ctx context.Context, rt *domainauth.RefreshToken) {
			rt.ID = 102
			rtSalvo = rt
		}).
		Return(nil)

	out, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "m@a.com", Senha: "senha123"})

	require.NoError(t, err)
	claims, err := jwtAuth.ValidarAccessToken(out.AccessToken)
	require.NoError(t, err)
	require.Equal(t, shared.PapelMecanico, claims.Papel)
	require.NotNil(t, rtSalvo)
	require.Equal(t, shared.PapelMecanico, rtSalvo.Papel)
	require.Equal(t, rtSalvo.AccessTokenJti, claims.Jti)
}

func TestLoginUseCase_Executar_SenhaErrada_ErroGenerico(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, mocks.NewJWTProvider(t), domainauth.TipoCliente, time.Hour, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: 1, SenhaHash: hashSenha(t, "senha123"), Ativo: true}, nil)

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "errada"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credenciais inválidas")
}

func TestLoginUseCase_Executar_UsuarioInativo_ErroGenerico(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, mocks.NewJWTProvider(t), domainauth.TipoCliente, time.Hour, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: 1, SenhaHash: hashSenha(t, "senha123"), Ativo: false}, nil)

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credenciais inválidas")
}

func TestLoginUseCase_Executar_ErroAoGerarAccessToken_RetornaErroInterno(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoCliente, time.Hour, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: 1, SenhaHash: hashSenha(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}, nil)
	jwtAuth.EXPECT().
		GerarAccessToken("1", domainauth.TipoCliente, shared.PapelCliente).
		Return("", "", errors.New("chave de assinatura invalida"))

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao gerar access token")
}

func TestLoginUseCase_Executar_ErroAoGerarRefreshToken_RetornaErroInterno(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoCliente, time.Hour, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: 1, SenhaHash: hashSenha(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}, nil)
	jwtAuth.EXPECT().
		GerarAccessToken("1", domainauth.TipoCliente, shared.PapelCliente).
		Return("access-token-valido", "jti-1", nil)
	jwtAuth.EXPECT().
		GerarRefreshToken().
		Return("", errors.New("entropia insuficiente"))

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao gerar refresh token")
}

func TestLoginUseCase_Executar_ErroAoSalvarRefreshToken_RetornaErroInterno(t *testing.T) {
	credenciaisRepo := mocks.NewCredenciaisRepository(t)
	refreshTokensRepo := mocks.NewRefreshTokenRepository(t)
	jwtAuth := mocks.NewJWTProvider(t)
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, jwtAuth, domainauth.TipoCliente, time.Hour, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "a@a.com").
		Return(&domainauth.Credencial{ID: 1, SenhaHash: hashSenha(t, "senha123"), Papel: shared.PapelCliente, Ativo: true}, nil)
	jwtAuth.EXPECT().
		GerarAccessToken("1", domainauth.TipoCliente, shared.PapelCliente).
		Return("access-token-valido", "jti-1", nil)
	jwtAuth.EXPECT().
		GerarRefreshToken().
		Return("refresh-token-bruto-de-teste", nil)
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
	uc := appauth.NewLoginUseCase(credenciaisRepo, refreshTokensRepo, mocks.NewJWTProvider(t), domainauth.TipoCliente, time.Hour, time.Hour)

	credenciaisRepo.EXPECT().
		BuscarPorEmail(mock.Anything, "naoexiste@a.com").
		Return(nil, errors.New("nao encontrado"))

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "naoexiste@a.com", Senha: "x"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credenciais inválidas")
}
