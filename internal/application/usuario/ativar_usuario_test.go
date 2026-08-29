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

func TestAtivarUsuarioUseCase_Executar_AtivaUsuarioInativo(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtivarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	existente.Inativar()

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).
		Run(func(ctx context.Context, u *domainusuario.Usuario) { require.True(t, u.Ativo()) }).
		Return(nil)

	require.NoError(t, uc.Executar(context.Background(), 1))
}

func TestAtivarUsuarioUseCase_Executar_UsuarioNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtivarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	err := uc.Executar(context.Background(), 99)
	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
}

func TestAtivarUsuarioUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtivarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtivarUsuarioUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtivarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	existente.Inativar()

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(errors.New("conexao recusada"))

	err = uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
