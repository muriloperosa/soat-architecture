package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	mysqlauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/auth"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func setupGormMock(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	require.NoError(t, err)

	gdb, err := gorm.Open(mysql.New(mysql.Config{Conn: db, SkipInitializeWithVersion: true}), &gorm.Config{})
	require.NoError(t, err)

	return gdb, mock
}

func TestRepositorioRefreshToken_Salvar_ExecutaInsert(t *testing.T) {
	gdb, mock := setupGormMock(t)
	repo := mysqlauth.NewRepositorioRefreshToken(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `refresh_tokens`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	rt := &domainauth.RefreshToken{UsuarioID: "user-1", Tipo: domainauth.TipoCliente, Papel: domainauth.PapelCliente, TokenHash: "hash-1", ExpiraEm: time.Now().Add(time.Hour)}
	err := repo.Salvar(context.Background(), rt)

	require.NoError(t, err)
	require.NotEmpty(t, rt.ID, "Salvar deve preencher o ID gerado")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorioRefreshToken_BuscarPorHash_RetornaEntidade(t *testing.T) {
	gdb, mock := setupGormMock(t)
	repo := mysqlauth.NewRepositorioRefreshToken(gdb)

	expiraEm := time.Now().Add(time.Hour)
	rows := sqlmock.NewRows([]string{"id", "usuario_id", "tipo", "papel", "token_hash", "expira_em", "revogado_em"}).
		AddRow("rt-1", "user-1", "cliente", "cliente", "hash-1", expiraEm, nil)
	mock.ExpectQuery("SELECT \\* FROM `refresh_tokens`").WillReturnRows(rows)

	rt, err := repo.BuscarPorHash(context.Background(), "hash-1")

	require.NoError(t, err)
	require.Equal(t, "rt-1", rt.ID)
	require.Equal(t, domainauth.PapelCliente, rt.Papel)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorioRefreshToken_Revogar_ExecutaUpdate(t *testing.T) {
	gdb, mock := setupGormMock(t)
	repo := mysqlauth.NewRepositorioRefreshToken(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `refresh_tokens`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repo.Revogar(context.Background(), "rt-1")

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
