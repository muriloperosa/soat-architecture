package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	mysqlauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/auth"
	test_helpers "github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func TestRefreshTokenRepository_Salvar_ExecutaInsert(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `refresh_tokens`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rt := &domainauth.RefreshToken{
		UsuarioID: 1, Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente,
		TokenHash: "hash-1", AccessTokenJti: "jti-1", ExpiraEm: time.Now().Add(time.Hour),
	}
	err := repo.Salvar(context.Background(), rt)

	require.NoError(t, err)
	require.Equal(t, uint64(1), rt.ID, "Salvar deve preencher o ID gerado (autoincrement)")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_Salvar_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `refresh_tokens`").WillReturnError(errors.New("conexao recusada"))
	mock.ExpectRollback()

	rt := &domainauth.RefreshToken{
		UsuarioID: 1, Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente,
		TokenHash: "hash-1", AccessTokenJti: "jti-1", ExpiraEm: time.Now().Add(time.Hour),
	}
	err := repo.Salvar(context.Background(), rt)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_Salvar_AtualizaRegistroExistente(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `refresh_tokens`").WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectCommit()

	rt := &domainauth.RefreshToken{
		ID: 7, UsuarioID: 1, Tipo: domainauth.TipoCliente, Papel: shared.PapelCliente,
		TokenHash: "hash-1", AccessTokenJti: "jti-1", ExpiraEm: time.Now().Add(time.Hour),
	}
	err := repo.Salvar(context.Background(), rt)

	require.NoError(t, err)
	require.Equal(t, uint64(7), rt.ID)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_BuscarPorHash_NaoEncontrado_RetornaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `refresh_tokens`").WillReturnError(gorm.ErrRecordNotFound)

	rt, err := repo.BuscarPorHash(context.Background(), "hash-inexistente")

	require.ErrorIs(t, err, domainauth.ErrRefreshTokenNaoEncontrado)
	require.Nil(t, rt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_BuscarPorHash_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `refresh_tokens`").WillReturnError(errors.New("conexao recusada"))

	rt, err := repo.BuscarPorHash(context.Background(), "hash-1")

	require.Error(t, err)
	require.Nil(t, rt)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_BuscarPorHash_RetornaEntidade(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	expiraEm := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"id", "usuario_id", "tipo", "papel", "token_hash", "access_token_jti", "expira_em", "revogado_em"}).
		AddRow(1, 1, "cliente", "CLIENTE", "hash-1", "jti-1", expiraEm, nil)
	mock.ExpectQuery("SELECT \\* FROM `refresh_tokens`").WillReturnRows(rows)

	rt, err := repo.BuscarPorHash(context.Background(), "hash-1")

	require.NoError(t, err)
	require.Equal(t, uint64(1), rt.ID)
	require.Equal(t, shared.PapelCliente, rt.Papel)
	require.Equal(t, "jti-1", rt.AccessTokenJti)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_Revogar_ExecutaUpdate(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `refresh_tokens`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Revogar(context.Background(), 1)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_Revogar_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `refresh_tokens`").WillReturnError(errors.New("conexao recusada"))
	mock.ExpectRollback()

	err := repo.Revogar(context.Background(), 1)

	require.Error(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_AccessTokenRevogado_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `refresh_tokens`").WillReturnError(errors.New("conexao recusada"))

	revogado, err := repo.AccessTokenRevogado(context.Background(), "jti-1")

	require.Error(t, err)
	require.False(t, revogado)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_AccessTokenRevogado_RetornaTrueQuandoExisteRevogado(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(1)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `refresh_tokens`").WillReturnRows(rows)

	revogado, err := repo.AccessTokenRevogado(context.Background(), "jti-1")

	require.NoError(t, err)
	require.True(t, revogado)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRefreshTokenRepository_AccessTokenRevogado_RetornaFalseQuandoNaoExiste(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlauth.NewRefreshTokenRepository(gdb)

	rows := sqlmock.NewRows([]string{"count"}).AddRow(0)
	mock.ExpectQuery("SELECT count\\(\\*\\) FROM `refresh_tokens`").WillReturnRows(rows)

	revogado, err := repo.AccessTokenRevogado(context.Background(), "jti-inexistente")

	require.NoError(t, err)
	require.False(t, revogado)
	require.NoError(t, mock.ExpectationsWereMet())
}
