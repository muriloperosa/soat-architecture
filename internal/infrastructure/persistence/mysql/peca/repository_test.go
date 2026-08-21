package peca

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newRepositoryTestDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
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
		&gorm.Config{
			SkipDefaultTransaction: true,
		},
	)
	require.NoError(t, err)

	return db, mock
}

func novaPecaValida(t *testing.T) *domain.Peca {
	t.Helper()

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	return p
}

func pecaRows() *sqlmock.Rows {
	agora := time.Now()

	return sqlmock.NewRows([]string{
		"id",
		"codigo",
		"nome",
		"marca",
		"descricao",
		"preco",
		"quantidade_em_estoque",
		"estoque_minimo",
		"criado_por",
		"ativo",
		"data_cadastro",
		"data_atualizacao",
	}).AddRow(
		uint64(1),
		"PC123",
		"Peca 1",
		"Marca 1",
		"Descricao 1",
		100.0,
		10,
		5,
		uint64(1),
		true,
		agora,
		agora,
	)
}

func TestNewRepository(t *testing.T) {
	db, _ := newRepositoryTestDB(t)

	repository := NewRepository(db)

	require.NotNil(t, repository)
	require.IsType(t, &Repository{}, repository)
}

func TestRepositorySalvar(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)

	mock.ExpectExec("INSERT INTO .*").WillReturnResult(sqlmock.NewResult(1, 1))

	err := repository.Salvar(context.Background(), p)

	require.NoError(t, err)
	require.Equal(t, uint64(1), p.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySalvarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)

	erroBanco := errors.New("erro ao inserir peca")

	mock.ExpectExec("INSERT INTO .*").WillReturnError(erroBanco)

	err := repository.Salvar(context.Background(), p)

	require.ErrorIs(t, err, erroBanco)
	require.Equal(t, uint64(0), p.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorID(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(pecaRows())

	p, err := repository.BuscarPorID(context.Background(), 1)

	require.NoError(t, err)
	require.NotNil(t, p)

	require.Equal(t, uint64(1), p.ID())
	require.Equal(t, "PC123", p.Codigo())
	require.Equal(t, "Peca 1", p.Nome())
	require.Equal(t, "Marca 1", p.Marca())
	require.Equal(t, 100.0, p.Preco())
	require.Equal(t, 10, p.QuantidadeEmEstoque())
	require.Equal(t, 5, p.EstoqueMinimo())
	require.True(t, p.Ativo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDDeveRetornarPecaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(gorm.ErrRecordNotFound)

	p, err := repository.BuscarPorID(context.Background(), 999)

	require.Nil(t, p)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	p, err := repository.BuscarPorID(context.Background(), 1)

	require.Nil(t, p)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorCodigo(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(pecaRows())

	p, err := repository.BuscarPorCodigo(context.Background(), "PC123")

	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, "PC123", p.Codigo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorCodigoDeveRetornarPecaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(gorm.ErrRecordNotFound)

	p, err := repository.BuscarPorCodigo(context.Background(), "INEXISTENTE")

	require.Nil(t, p)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorCodigoDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	p, err := repository.BuscarPorCodigo(context.Background(), "PC123")

	require.Nil(t, p)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizar(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)
	p.AtribuirID(1)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Atualizar(context.Background(), p)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarPecaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)
	p.AtribuirID(999)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.Atualizar(context.Background(), p)

	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)
	p.AtribuirID(1)

	erroBanco := errors.New("erro ao atualizar peca")

	mock.ExpectExec("UPDATE .*").WillReturnError(erroBanco)

	err := repository.Atualizar(context.Background(), p)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}
