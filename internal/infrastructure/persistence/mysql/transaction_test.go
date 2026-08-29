package mysql

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New()
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	db, err := gorm.Open(
		gormmysql.New(gormmysql.Config{
			Conn:                      sqlDB,
			SkipInitializeWithVersion: true,
		}),
		&gorm.Config{},
	)
	require.NoError(t, err)

	return db, mock
}

func TestDBFromContext_SemTransacaoNoContexto_RetornaDB(t *testing.T) {
	db, _ := newTestDB(t)

	require.Same(t, db, DBFromContext(context.Background(), db))
}

func TestDBFromContext_ComTransacaoNoContexto_RetornaTx(t *testing.T) {
	db, _ := newTestDB(t)
	tx := db.Session(&gorm.Session{})

	ctx := context.WithValue(context.Background(), txContextKey{}, tx)

	require.Same(t, tx, DBFromContext(ctx, db))
}

func TestTransactionRunner_Executar_ComSucesso_Commita(t *testing.T) {
	db, mock := newTestDB(t)
	runner := NewTransactionRunner(db)

	mock.ExpectBegin()
	mock.ExpectCommit()

	var ctxRecebido context.Context
	err := runner.Executar(context.Background(), func(ctx context.Context) error {
		ctxRecebido = ctx
		return nil
	})

	require.NoError(t, err)
	require.NotNil(t, ctxRecebido)
	require.NotSame(t, db, DBFromContext(ctxRecebido, db), "fn deveria receber um ctx com a transação injetada")
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestTransactionRunner_Executar_ComErro_FazRollback(t *testing.T) {
	db, mock := newTestDB(t)
	runner := NewTransactionRunner(db)

	erroFn := errors.New("erro dentro da transação")

	mock.ExpectBegin()
	mock.ExpectRollback()

	err := runner.Executar(context.Background(), func(ctx context.Context) error {
		return erroFn
	})

	require.ErrorIs(t, err, erroFn)
	require.NoError(t, mock.ExpectationsWereMet())
}
