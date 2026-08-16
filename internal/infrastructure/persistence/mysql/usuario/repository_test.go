package usuario_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	mysqlusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/usuario"
	test_helpers "github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestUsuarioRepository_Salvar_ExecutaInsertEPreencheID(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `usuarios`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	u, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = repo.Salvar(context.Background(), u)
	require.NoError(t, err)
	require.Equal(t, uint64(1), u.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_Salvar_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `usuarios`").WillReturnError(errors.New("conexao recusada"))
	mock.ExpectRollback()

	u, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)

	err = repo.Salvar(context.Background(), u)
	require.Error(t, err)
	require.Zero(t, u.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_BuscarPorID_NaoEncontrado_RetornaErrUsuarioNaoEncontrado(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `usuarios`").WillReturnError(gorm.ErrRecordNotFound)

	u, err := repo.BuscarPorID(context.Background(), 99)
	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
	require.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_BuscarPorID_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `usuarios`").WillReturnError(errors.New("conexao recusada"))

	u, err := repo.BuscarPorID(context.Background(), 1)
	require.Error(t, err)
	require.False(t, errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado))
	require.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_BuscarPorEmail_NaoEncontrado_RetornaErrUsuarioNaoEncontrado(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `usuarios`").WillReturnError(gorm.ErrRecordNotFound)

	u, err := repo.BuscarPorEmail(context.Background(), "naoexiste@oficina.com")
	require.ErrorIs(t, err, domainusuario.ErrUsuarioNaoEncontrado)
	require.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_BuscarPorEmail_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `usuarios`").WillReturnError(errors.New("conexao recusada"))

	u, err := repo.BuscarPorEmail(context.Background(), "ana@oficina.com")
	require.Error(t, err)
	require.False(t, errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado))
	require.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_BuscarPorEmail_RetornaEntidade(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	agora := time.Now()
	rows := sqlmock.NewRows([]string{"id", "papel", "nome", "email", "senha_hash", "requer_alterar_senha", "ativo", "data_cadastro", "data_atualizacao"}).
		AddRow(1, "MECANICO", "Ana Souza", "ana@oficina.com", "$2a$hash", false, true, agora, agora)
	mock.ExpectQuery("SELECT \\* FROM `usuarios`").WillReturnRows(rows)

	u, err := repo.BuscarPorEmail(context.Background(), "ana@oficina.com")
	require.NoError(t, err)
	require.Equal(t, uint64(1), u.ID())
	require.Equal(t, "ana@oficina.com", u.Email().String())
	require.Equal(t, shared.PapelMecanico, u.Papel())
	require.False(t, u.RequerAlterarSenha())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_BuscarPorID_RetornaEntidade(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	agora := time.Now()
	rows := sqlmock.NewRows([]string{"id", "papel", "nome", "email", "senha_hash", "requer_alterar_senha", "ativo", "data_cadastro", "data_atualizacao"}).
		AddRow(1, "MECANICO", "Ana Souza", "ana@oficina.com", "$2a$hash", true, true, agora, agora)
	mock.ExpectQuery("SELECT \\* FROM `usuarios`").WillReturnRows(rows)

	u, err := repo.BuscarPorID(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), u.ID())
	require.Equal(t, shared.PapelMecanico, u.Papel())
	require.True(t, u.RequerAlterarSenha())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_BuscarPorID_EmailInvalidoNoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	agora := time.Now()
	rows := sqlmock.NewRows([]string{"id", "papel", "nome", "email", "senha_hash", "requer_alterar_senha", "ativo", "data_cadastro", "data_atualizacao"}).
		AddRow(1, "MECANICO", "Ana Souza", "email-corrompido-no-banco", "$2a$hash", true, true, agora, agora)
	mock.ExpectQuery("SELECT \\* FROM `usuarios`").WillReturnRows(rows)

	u, err := repo.BuscarPorID(context.Background(), 1)
	require.ErrorIs(t, err, shared.ErrEmailInvalido)
	require.Nil(t, u)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestUsuarioRepository_Atualizar_ExecutaUpdate(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlusuario.NewUsuarioRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `usuarios`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	u, err := domainusuario.NewUsuario("Ana Souza", "ana@oficina.com", "senha123", shared.PapelMecanico)
	require.NoError(t, err)
	u.AtribuirID(1)

	err = repo.Atualizar(context.Background(), u)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
