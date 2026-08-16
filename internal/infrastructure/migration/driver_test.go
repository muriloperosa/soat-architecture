package migration

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"

	db "github.com/muriloperosa/soat-architecture/internal/infrastructure/database"
)

const (
	databaseName = "mecanica"
)

func TestNewDriverMySQL(t *testing.T) {
	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	mock.ExpectQuery(`SELECT GET_LOCK\(\?, 10\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnRows(
			sqlmock.NewRows([]string{"GET_LOCK"}).AddRow(true),
		)

	mock.ExpectQuery(`SHOW TABLES LIKE 'schema_migrations'`).WillReturnError(sql.ErrNoRows)

	mock.ExpectExec(
		"CREATE TABLE `schema_migrations` " +
			"\\(version bigint not null primary key, dirty boolean not null\\)",
	).WillReturnResult(sqlmock.NewResult(0, 0))

	mock.ExpectExec(`SELECT RELEASE_LOCK\(\?\)`).
		WithArgs(sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	driver, err := NewDriver(db.DriverTypeMySQL, sqlDB, databaseName)

	require.NoError(t, err)
	require.NotNil(t, driver)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewDriver_DeveRetornarErroAoCriarDriverMySQL(t *testing.T) {
	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer sqlDB.Close()

	mock.ExpectPing().WillReturnError(errors.New("erro de conexão"))

	driver, err := NewDriver(db.DriverTypeMySQL, sqlDB, "test_database")

	require.Nil(t, driver)
	require.Error(t, err)
	require.Contains(t, err.Error(), "erro ao criar driver MySQL")
	require.Contains(t, err.Error(), "erro de conexão")

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestNewDriverComTipoInvalido(t *testing.T) {
	driver, err := NewDriver(db.DriverType("mongodb"), nil, databaseName)

	require.Nil(t, driver)
	require.ErrorIs(t, err, ErrDriverInvalido)
}

func TestNewDriverMySQLSemConexao(t *testing.T) {
	driver, err := NewDriver(db.DriverTypeMySQL, nil, databaseName)

	require.Nil(t, driver)
	require.ErrorIs(t, err, ErrConexaoObrigatoria)
}

func TestNewDriverComDriversAindaNaoSuportados(t *testing.T) {
	tests := []struct {
		name   string
		driver db.DriverType
	}{
		{
			name:   "postgresql",
			driver: db.DriverTypePostgreSQL,
		},
		{
			name:   "sqlite",
			driver: db.DriverTypeSQLite,
		},
		{
			name:   "sql server",
			driver: db.DriverTypeSQLServer,
		},
		{
			name:   "oracle",
			driver: db.DriverTypeOracle,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := NewDriver(tt.driver, nil, databaseName)

			require.Nil(t, driver)
			require.ErrorIs(t, err, ErrDriverNaoSuportado)
		})
	}
}
