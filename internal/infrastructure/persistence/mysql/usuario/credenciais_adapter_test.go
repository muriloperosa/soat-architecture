package usuario_test

import (
	"context"
	"errors"
	"testing"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/usuario/mocks"
	mysqlusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/usuario"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	_ domainauth.CredenciaisRepository   = (*mysqlusuario.CredenciaisAdapter)(nil)
	_ domainauth.UsuarioStatusRepository = (*mysqlusuario.CredenciaisAdapter)(nil)
)

func TestCredenciaisAdapter_BuscarPorEmail_MapeiaUsuarioParaCredencial(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	adapter := mysqlusuario.NewCredenciaisAdapter(repo)

	u, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	u.AtribuirID(1)

	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(u, nil)

	cred, err := adapter.BuscarPorEmail(context.Background(), "ana@oficina.com")
	require.NoError(t, err)
	require.Equal(t, uint64(1), cred.ID)
	require.Equal(t, shared.PapelMecanico, cred.Papel)
	require.True(t, cred.Ativo)
	require.True(t, cred.RequerAlterarSenha)
}

func TestCredenciaisAdapter_BuscarPorEmail_NaoEncontrado_PropagaErro(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	adapter := mysqlusuario.NewCredenciaisAdapter(repo)

	repo.EXPECT().BuscarPorEmail(mock.Anything, "naoexiste@oficina.com").Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	_, err := adapter.BuscarPorEmail(context.Background(), "naoexiste@oficina.com")
	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
}

func TestCredenciaisAdapter_EstaAtivo_UsuarioAtivo_RetornaTrue(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	adapter := mysqlusuario.NewCredenciaisAdapter(repo)

	u, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	u.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(u, nil)

	ativo, err := adapter.EstaAtivo(context.Background(), 1)
	require.NoError(t, err)
	require.True(t, ativo)
}

func TestCredenciaisAdapter_EstaAtivo_UsuarioInativo_RetornaFalse(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	adapter := mysqlusuario.NewCredenciaisAdapter(repo)

	u, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	u.AtribuirID(1)
	u.Inativar()

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(u, nil)

	ativo, err := adapter.EstaAtivo(context.Background(), 1)
	require.NoError(t, err)
	require.False(t, ativo)
}

func TestCredenciaisAdapter_EstaAtivo_UsuarioNaoEncontrado_PropagaErro(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	adapter := mysqlusuario.NewCredenciaisAdapter(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	ativo, err := adapter.EstaAtivo(context.Background(), 99)
	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
	require.False(t, ativo)
}

func TestCredenciaisAdapter_EstaAtivo_ErroDoBanco_PropagaErro(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	adapter := mysqlusuario.NewCredenciaisAdapter(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	ativo, err := adapter.EstaAtivo(context.Background(), 1)
	require.Error(t, err)
	require.False(t, ativo)
}
