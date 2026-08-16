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

func TestAtualizarUsuarioUseCase_Executar_UsuarioExiste_AtualizaNomeEPapel(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(nil)

	out, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana S. Costa", Papel: shared.PapelAtendente})

	require.NoError(t, err)
	require.Equal(t, "Ana S. Costa", out.Nome)
	require.Equal(t, shared.PapelAtendente, out.Papel)
}

func TestAtualizarUsuarioUseCase_Executar_UsuarioNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	_, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 99, Nome: "X", Papel: shared.PapelAdmin})

	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
}

func TestAtualizarUsuarioUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana S. Costa", Papel: shared.PapelAtendente})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtualizarUsuarioUseCase_Executar_DadosInvalidos_PropagaErroDeValidacao(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "", Papel: shared.PapelAtendente})

	require.ErrorIs(t, err, domainusuario.ErrNomeObrigatorio)
}

func TestAtualizarUsuarioUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(errors.New("conexao recusada"))

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana S. Costa", Papel: shared.PapelAtendente})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
