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

func TestAlterarSenhaUseCase_Executar_EncerraRequerAlterarSenha(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAlterarSenhaUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	require.True(t, existente.RequerAlterarSenha())

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).
		Run(func(ctx context.Context, u *domainusuario.Usuario) {
			require.False(t, u.RequerAlterarSenha())
			require.True(t, u.Senha().Confere("senhaNova123"))
		}).
		Return(nil)

	err = uc.Executar(context.Background(), appusuario.AlterarSenhaInput{UsuarioID: 1, SenhaNova: "senhaNova123"})
	require.NoError(t, err)
}

func TestAlterarSenhaUseCase_Executar_SenhaFraca_RetornaErro(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAlterarSenhaUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	err = uc.Executar(context.Background(), appusuario.AlterarSenhaInput{UsuarioID: 1, SenhaNova: "curta"})
	require.ErrorIs(t, err, shared.ErrSenhaFraca)
}

func TestAlterarSenhaUseCase_Executar_UsuarioNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAlterarSenhaUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	err := uc.Executar(context.Background(), appusuario.AlterarSenhaInput{UsuarioID: 99, SenhaNova: "senhaNova123"})
	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
}

func TestAlterarSenhaUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAlterarSenhaUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), appusuario.AlterarSenhaInput{UsuarioID: 1, SenhaNova: "senhaNova123"})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAlterarSenhaUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAlterarSenhaUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(errors.New("conexao recusada"))

	err = uc.Executar(context.Background(), appusuario.AlterarSenhaInput{UsuarioID: 1, SenhaNova: "senhaNova123"})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
