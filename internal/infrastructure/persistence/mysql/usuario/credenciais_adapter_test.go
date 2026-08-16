package usuario_test

import (
	"context"
	"testing"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/usuario/mocks"
	mysqlusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/usuario"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var _ domainauth.CredenciaisRepository = (*mysqlusuario.CredenciaisAdapter)(nil)

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
