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

	out, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana S. Costa", Email: "ana@oficina.com", Papel: string(shared.PapelAtendente)})

	require.NoError(t, err)
	require.Equal(t, "Ana S. Costa", out.Nome)
	require.Equal(t, shared.PapelAtendente, out.Papel)
}

func TestAtualizarUsuarioUseCase_Executar_UsuarioNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	_, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 99, Nome: "X", Email: "x@oficina.com", Papel: string(shared.PapelAdmin)})

	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
}

func TestAtualizarUsuarioUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana S. Costa", Email: "ana@oficina.com", Papel: string(shared.PapelAtendente)})

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

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "", Email: "ana@oficina.com", Papel: string(shared.PapelAtendente)})

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

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana S. Costa", Email: "ana@oficina.com", Papel: string(shared.PapelAtendente)})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtualizarUsuarioUseCase_Executar_TrocaEmailParaUmLivre_Atualiza(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana.nova@oficina.com").Return(nil, domainusuario.ErrUsuarioNaoEncontrado)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(nil)

	out, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana Souza", Email: "ana.nova@oficina.com", Papel: string(shared.PapelMecanico)})

	require.NoError(t, err)
	require.Equal(t, "ana.nova@oficina.com", out.Email)
}

func TestAtualizarUsuarioUseCase_Executar_TrocaEmailJaUsadoPorOutroUsuario_RetornaConflict(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	outroUsuario, err := domainusuario.NewUsuario("Bia Lima", "bia@oficina.com", "senha123", shared.PapelAtendente)
	require.NoError(t, err)
	outroUsuario.AtribuirID(2)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().BuscarPorEmail(mock.Anything, "bia@oficina.com").Return(outroUsuario, nil)

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana Souza", Email: "bia@oficina.com", Papel: string(shared.PapelMecanico)})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindConflict, appErr.Kind)
}

func TestAtualizarUsuarioUseCase_Executar_ErroDoBancoAoVerificarEmail_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana.nova@oficina.com").Return(nil, errors.New("conexao recusada"))

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana Souza", Email: "ana.nova@oficina.com", Papel: string(shared.PapelMecanico)})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtualizarUsuarioUseCase_Executar_ComSenhaNova_RedefineEForcaTroca(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	require.NoError(t, existente.AlterarSenha("senhaAntiga123"))
	require.False(t, existente.RequerAlterarSenha())

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).
		Run(func(ctx context.Context, u *domainusuario.Usuario) {
			require.True(t, u.RequerAlterarSenha())
			require.True(t, u.Senha().Confere("senhaDoAdmin123"))
		}).
		Return(nil)

	out, err := uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana Souza", Email: "ana@oficina.com", SenhaNova: "senhaDoAdmin123", Papel: string(shared.PapelMecanico)})

	require.NoError(t, err)
	require.True(t, out.RequerAlterarSenha)
}

func TestAtualizarUsuarioUseCase_Executar_SemSenhaNova_NaoAlteraSenha(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)
	require.NoError(t, existente.AlterarSenha("senhaAntiga123"))

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).
		Run(func(ctx context.Context, u *domainusuario.Usuario) {
			require.False(t, u.RequerAlterarSenha())
			require.True(t, u.Senha().Confere("senhaAntiga123"))
		}).
		Return(nil)

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana Souza", Email: "ana@oficina.com", Papel: string(shared.PapelMecanico)})
	require.NoError(t, err)
}

func TestAtualizarUsuarioUseCase_Executar_SenhaNovaFraca_RetornaErro(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewAtualizarUsuarioUseCase(repo)

	existente, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	_, err = uc.Executar(context.Background(), appusuario.AtualizarUsuarioInput{ID: 1, Nome: "Ana Souza", Email: "ana@oficina.com", SenhaNova: "curta", Papel: string(shared.PapelMecanico)})

	require.ErrorIs(t, err, shared.ErrSenhaFraca)
}
