package reservapeca

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
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

func novaReservaValida(t *testing.T) *domain.ReservaPeca {
	t.Helper()

	r, err := domain.NewReservaPeca(1, 2, 3)
	require.NoError(t, err)

	return r
}

func reservaRows() *sqlmock.Rows {
	agora := time.Now()

	return sqlmock.NewRows([]string{
		"id",
		"ordem_servico_id",
		"peca_id",
		"quantidade",
		"criada_em",
		"atualizada_em",
	}).AddRow(uint64(1), uint64(1), uint64(2), 3, agora, agora)
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

	r := novaReservaValida(t)

	mock.ExpectExec("INSERT INTO .*").WillReturnResult(sqlmock.NewResult(1, 1))

	err := repository.Salvar(context.Background(), r)

	require.NoError(t, err)
	require.Equal(t, uint64(1), r.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySalvar_ParticipaDaTransacaoDoContexto(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)
	runner := mysql.NewTransactionRunner(db)

	r := novaReservaValida(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO .*").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	err := runner.Executar(context.Background(), func(ctx context.Context) error {
		return repository.Salvar(ctx, r)
	})

	require.NoError(t, err)
	require.Equal(t, uint64(1), r.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySalvarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	r := novaReservaValida(t)

	erroBanco := errors.New("erro ao inserir reserva")

	mock.ExpectExec("INSERT INTO .*").WillReturnError(erroBanco)

	err := repository.Salvar(context.Background(), r)

	require.ErrorIs(t, err, erroBanco)
	require.Equal(t, uint64(0), r.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizar(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	r := novaReservaValida(t)
	r.AtribuirID(1)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Atualizar(context.Background(), r)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarReservaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	r := novaReservaValida(t)
	r.AtribuirID(999)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.Atualizar(context.Background(), r)

	require.ErrorIs(t, err, domain.ErrReservaNaoEncontrada)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	r := novaReservaValida(t)
	r.AtribuirID(1)

	erroBanco := errors.New("erro ao atualizar reserva")

	mock.ExpectExec("UPDATE .*").WillReturnError(erroBanco)

	err := repository.Atualizar(context.Background(), r)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemEPeca(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(reservaRows())

	r, err := repository.BuscarPorOrdemEPeca(context.Background(), 1, 2)

	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, uint64(1), r.ID())
	require.Equal(t, uint64(1), r.OrdemServicoID())
	require.Equal(t, uint64(2), r.PecaID())
	require.Equal(t, 3, r.Quantidade())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemEPecaDeveRetornarReservaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(gorm.ErrRecordNotFound)

	r, err := repository.BuscarPorOrdemEPeca(context.Background(), 1, 999)

	require.Nil(t, r)
	require.ErrorIs(t, err, domain.ErrReservaNaoEncontrada)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemEPecaDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	r, err := repository.BuscarPorOrdemEPeca(context.Background(), 1, 2)

	require.Nil(t, r)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemEPecaComBloqueio(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .* FOR UPDATE").WillReturnRows(reservaRows())

	r, err := repository.BuscarPorOrdemEPecaComBloqueio(context.Background(), 1, 2)

	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, uint64(1), r.ID())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemEPecaComBloqueioDeveRetornarReservaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .* FOR UPDATE").WillReturnError(gorm.ErrRecordNotFound)

	r, err := repository.BuscarPorOrdemEPecaComBloqueio(context.Background(), 1, 999)

	require.Nil(t, r)
	require.ErrorIs(t, err, domain.ErrReservaNaoEncontrada)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemEPecaComBloqueio_ParticipaDaTransacaoDoContexto(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)
	runner := mysql.NewTransactionRunner(db)

	mock.ExpectBegin()
	mock.ExpectQuery("SELECT .* FROM .* FOR UPDATE").WillReturnRows(reservaRows())
	mock.ExpectCommit()

	err := runner.Executar(context.Background(), func(ctx context.Context) error {
		r, err := repository.BuscarPorOrdemEPecaComBloqueio(ctx, 1, 2)
		require.NoError(t, err)
		require.NotNil(t, r)
		return nil
	})

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemServico(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(reservaRows())

	reservas, err := repository.BuscarPorOrdemServico(context.Background(), 1)

	require.NoError(t, err)
	require.Len(t, reservas, 1)
	require.Equal(t, uint64(2), reservas[0].PecaID())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemServicoDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	reservas, err := repository.BuscarPorOrdemServico(context.Background(), 1)

	require.Nil(t, reservas)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySomarQuantidadeReservada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	rows := sqlmock.NewRows([]string{"COALESCE(SUM(quantidade), 0)"}).AddRow(7)
	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(rows)

	total, err := repository.SomarQuantidadeReservada(context.Background(), 2)

	require.NoError(t, err)
	require.Equal(t, 7, total)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySomarQuantidadeReservadaDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	total, err := repository.SomarQuantidadeReservada(context.Background(), 2)

	require.Zero(t, total)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryRemover(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectExec("DELETE FROM .*").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Remover(context.Background(), 1, 2)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryRemoverDeveRetornarReservaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectExec("DELETE FROM .*").WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.Remover(context.Background(), 1, 999)

	require.ErrorIs(t, err, domain.ErrReservaNaoEncontrada)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryRemoverDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao remover reserva")

	mock.ExpectExec("DELETE FROM .*").WillReturnError(erroBanco)

	err := repository.Remover(context.Background(), 1, 2)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}
