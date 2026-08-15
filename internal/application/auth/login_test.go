package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type credenciaisFake struct {
	credencial *domainauth.Credencial
}

func (f *credenciaisFake) BuscarPorEmail(ctx context.Context, email string) (*domainauth.Credencial, error) {
	if f.credencial == nil {
		return nil, errors.New("nao encontrado")
	}
	return f.credencial, nil
}

type refreshTokensFake struct {
	salvos []*domainauth.RefreshToken
}

func (f *refreshTokensFake) Salvar(ctx context.Context, rt *domainauth.RefreshToken) error {
	rt.ID = "rt-1"
	f.salvos = append(f.salvos, rt)
	return nil
}
func (f *refreshTokensFake) BuscarPorHash(ctx context.Context, hash string) (*domainauth.RefreshToken, error) {
	for _, rt := range f.salvos {
		if rt.TokenHash == hash {
			return rt, nil
		}
	}
	return nil, errors.New("nao encontrado")
}
func (f *refreshTokensFake) Revogar(ctx context.Context, id string) error {
	for _, rt := range f.salvos {
		if rt.ID == id {
			now := time.Now()
			rt.RevogadoEm = &now
		}
	}
	return nil
}

func hashSenha(t *testing.T, senha string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(senha), bcrypt.DefaultCost)
	require.NoError(t, err)
	return string(h)
}

func TestLoginUseCase_Executar_CredencialValida_RetornaTokens(t *testing.T) {
	credenciais := &credenciaisFake{credencial: &domainauth.Credencial{ID: "user-1", SenhaHash: hashSenha(t, "senha123"), Papel: domainauth.PapelCliente}}
	refreshTokens := &refreshTokensFake{}
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)
	uc := appauth.NewLoginUseCase(credenciais, refreshTokens, jwtAuth, domainauth.TipoCliente, 168*time.Hour)

	out, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "senha123"})

	require.NoError(t, err)
	require.NotEmpty(t, out.AccessToken)
	require.NotEmpty(t, out.RefreshToken)
	require.Len(t, refreshTokens.salvos, 1)
	require.Equal(t, "user-1", refreshTokens.salvos[0].UsuarioID)
	require.Equal(t, domainauth.TipoCliente, refreshTokens.salvos[0].Tipo)
	require.Equal(t, domainauth.PapelCliente, refreshTokens.salvos[0].Papel)

	claims, err := jwtAuth.ValidarAccessToken(out.AccessToken)
	require.NoError(t, err)
	require.Equal(t, domainauth.PapelCliente, claims.Papel)
}

func TestLoginUseCase_Executar_UsuarioInterno_PropagaPapel(t *testing.T) {
	credenciais := &credenciaisFake{credencial: &domainauth.Credencial{ID: "user-2", SenhaHash: hashSenha(t, "senha123"), Papel: domainauth.PapelMecanico}}
	refreshTokens := &refreshTokensFake{}
	jwtAuth := infraauth.NewAuthenticatorJWT("segredo-de-teste", 15*time.Minute)
	uc := appauth.NewLoginUseCase(credenciais, refreshTokens, jwtAuth, domainauth.TipoInterno, 168*time.Hour)

	out, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "m@a.com", Senha: "senha123"})

	require.NoError(t, err)
	claims, err := jwtAuth.ValidarAccessToken(out.AccessToken)
	require.NoError(t, err)
	require.Equal(t, domainauth.PapelMecanico, claims.Papel)
	require.Equal(t, domainauth.PapelMecanico, refreshTokens.salvos[0].Papel)
}

func TestLoginUseCase_Executar_SenhaErrada_ErroGenerico(t *testing.T) {
	credenciais := &credenciaisFake{credencial: &domainauth.Credencial{ID: "user-1", SenhaHash: hashSenha(t, "senha123")}}
	uc := appauth.NewLoginUseCase(credenciais, &refreshTokensFake{}, infraauth.NewAuthenticatorJWT("s", time.Minute), domainauth.TipoCliente, time.Hour)

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "a@a.com", Senha: "errada"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credenciais inválidas")
}

func TestLoginUseCase_Executar_EmailNaoEncontrado_MesmoErroGenerico(t *testing.T) {
	uc := appauth.NewLoginUseCase(&credenciaisFake{}, &refreshTokensFake{}, infraauth.NewAuthenticatorJWT("s", time.Minute), domainauth.TipoCliente, time.Hour)

	_, err := uc.Executar(context.Background(), appauth.LoginInput{Email: "naoexiste@a.com", Senha: "x"})

	require.Error(t, err)
	require.Contains(t, err.Error(), "credenciais inválidas")
}
