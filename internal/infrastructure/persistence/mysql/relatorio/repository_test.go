package relatorio

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newRepositoryTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{})
	require.NoError(t, err)

	return db, mock
}

func paramsDeTeste(t *testing.T) domain.CalcularTransicaoStatusParams {
	t.Helper()
	inicio := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	fim := time.Date(2026, time.March, 31, 0, 0, 0, 0, time.UTC)
	periodo, err := domain.NewPeriodo(inicio, fim)
	require.NoError(t, err)

	return domain.CalcularTransicaoStatusParams{
		FromStatus: ordemservico.StatusRecebida,
		ToStatus:   ordemservico.StatusEmDiagnostico,
		Periodo:    periodo,
	}
}

func TestRepositoryCalcularTransicaoStatusComResultado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRelatorioRepository(db)
	params := paramsDeTeste(t)

	mock.ExpectQuery("WITH primeiro_from AS").
		WithArgs(
			params.FromStatus.String(),
			params.ToStatus.String(),
			params.Periodo.Inicio(),
			params.Periodo.Fim(),
		).
		WillReturnRows(sqlmock.NewRows([]string{"total", "media_segundos", "minima_segundos", "maxima_segundos"}).
			AddRow(3, 7200.0, 3600.0, 10800.0))

	resultado, err := repository.CalcularTransicaoStatus(context.Background(), params)

	require.NoError(t, err)
	require.Equal(t, 3, resultado.TotalOrdens)
	require.Equal(t, 2*time.Hour, resultado.DuracaoMedia)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryCalcularTransicaoStatusSemResultado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRelatorioRepository(db)
	params := paramsDeTeste(t)

	mock.ExpectQuery("WITH primeiro_from AS").
		WillReturnRows(sqlmock.NewRows([]string{"total", "media_segundos", "minima_segundos", "maxima_segundos"}).
			AddRow(0, nil, nil, nil))

	resultado, err := repository.CalcularTransicaoStatus(context.Background(), params)

	require.NoError(t, err)
	require.Zero(t, resultado.TotalOrdens)
	require.Zero(t, resultado.DuracaoMedia)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryCalcularTransicaoStatusErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRelatorioRepository(db)
	params := paramsDeTeste(t)
	erroBanco := errors.New("conexao recusada")

	mock.ExpectQuery("WITH primeiro_from AS").WillReturnError(erroBanco)

	_, err := repository.CalcularTransicaoStatus(context.Background(), params)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}
