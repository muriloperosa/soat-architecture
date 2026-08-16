package usuario_test

import (
	"context"
	"errors"
	"testing"

	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/usuario/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestBuscarUsuarioLogadoUseCase_Executar_UsuarioExiste_RetornaDados(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewBuscarUsuarioLogadoUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	out, err := uc.Executar(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), out.ID)
	require.Equal(t, "ana@oficina.com", out.Email)
}

func TestBuscarUsuarioLogadoUseCase_Executar_UsuarioNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewBuscarUsuarioLogadoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	_, err := uc.Executar(context.Background(), 99)
	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
}

func TestBuscarUsuarioLogadoUseCase_Executar_ErroDoBanco_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewBuscarUsuarioLogadoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
