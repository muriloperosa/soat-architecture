package usuario_test

import (
	"context"
	"errors"
	"testing"
	"time"

	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	"github.com/muriloperosa/soat-architecture/internal/domain/usuario/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func mustEmail(t *testing.T, valor string) shared.Email {
	t.Helper()
	e, err := shared.NewEmail(valor)
	require.NoError(t, err)
	return e
}

func TestCriarUsuarioUseCase_Executar_EmailNovo_CriaComRequerAlterarSenha(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewCriarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(nil, domainusuario.ErrUsuarioNaoEncontrado)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).
		Run(func(ctx context.Context, u *domainusuario.Usuario) { u.AtribuirID(1) }).
		Return(nil)

	out, err := uc.Executar(context.Background(), appusuario.CriarUsuarioInput{
		Nome: "Ana Souza", Email: "ana@oficina.com", SenhaInicial: "senha123", Papel: shared.PapelMecanico,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(1), out.ID)
	require.True(t, out.RequerAlterarSenha)
	require.True(t, out.Ativo)
}

func TestCriarUsuarioUseCase_Executar_EmailJaExiste_RetornaConflict(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewCriarUsuarioUseCase(repo)

	existente := domainusuario.RestaurarUsuario(1, "Ana", mustEmail(t, "ana@oficina.com"), shared.RestaurarSenhaHash("h"), shared.PapelMecanico, false, true, time.Now(), time.Now())
	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(existente, nil)

	_, err := uc.Executar(context.Background(), appusuario.CriarUsuarioInput{
		Nome: "Ana Souza", Email: "ana@oficina.com", SenhaInicial: "senha123", Papel: shared.PapelMecanico,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindConflict, appErr.Kind)
}

func TestCriarUsuarioUseCase_Executar_ErroDoBancoAoVerificarEmail_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewCriarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appusuario.CriarUsuarioInput{
		Nome: "Ana Souza", Email: "ana@oficina.com", SenhaInicial: "senha123", Papel: shared.PapelMecanico,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestCriarUsuarioUseCase_Executar_DadosInvalidos_PropagaErroDeValidacao(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewCriarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(nil, domainusuario.ErrUsuarioNaoEncontrado)

	_, err := uc.Executar(context.Background(), appusuario.CriarUsuarioInput{
		Nome: "Ana Souza", Email: "ana@oficina.com", SenhaInicial: "curta", Papel: shared.PapelMecanico,
	})

	require.ErrorIs(t, err, shared.ErrSenhaFraca)
}

func TestCriarUsuarioUseCase_Executar_ErroDoBancoAoSalvar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewUsuarioRepository(t)
	uc := appusuario.NewCriarUsuarioUseCase(repo)

	repo.EXPECT().BuscarPorEmail(mock.Anything, "ana@oficina.com").Return(nil, domainusuario.ErrUsuarioNaoEncontrado)
	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*usuario.Usuario")).Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appusuario.CriarUsuarioInput{
		Nome: "Ana Souza", Email: "ana@oficina.com", SenhaInicial: "senha123", Papel: shared.PapelMecanico,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
