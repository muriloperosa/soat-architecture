package veiculo

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
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

func novoVeiculoValido(t *testing.T) *domain.Veiculo {
	t.Helper()

	v, err := domain.NewVeiculo("ABC1D23", "Fiat", "Uno", 15000, 2020, "Prata", 1)
	require.NoError(t, err)

	return v
}

func veiculoRows() *sqlmock.Rows {
	agora := time.Now()

	return sqlmock.NewRows([]string{
		"id",
		"placa",
		"marca",
		"modelo",
		"quilometragem_atual",
		"ano",
		"cor",
		"criado_por",
		"ativo",
		"data_cadastro",
		"data_atualizacao",
	}).AddRow(
		uint64(1),
		"ABC1D23",
		"Fiat",
		"Uno",
		uint32(15000),
		uint16(2020),
		"Prata",
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

	v := novoVeiculoValido(t)

	mock.ExpectExec("INSERT INTO .*").WillReturnResult(sqlmock.NewResult(1, 1))

	err := repository.Salvar(context.Background(), v)

	require.NoError(t, err)
	require.Equal(t, uint64(1), v.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySalvarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	v := novoVeiculoValido(t)

	erroBanco := errors.New("erro ao inserir veiculo")

	mock.ExpectExec("INSERT INTO .*").WillReturnError(erroBanco)

	err := repository.Salvar(context.Background(), v)

	require.ErrorIs(t, err, erroBanco)
	require.Equal(t, uint64(0), v.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorID(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(veiculoRows())

	v, err := repository.BuscarPorID(context.Background(), 1)

	require.NoError(t, err)
	require.NotNil(t, v)

	require.Equal(t, uint64(1), v.ID())
	require.Equal(t, "ABC1D23", v.Placa().String())
	require.Equal(t, "Fiat", v.Marca())
	require.Equal(t, "Uno", v.Modelo())
	require.Equal(t, uint32(15000), v.QuilometragemAtual())
	require.Equal(t, uint16(2020), v.Ano())
	require.Equal(t, "Prata", v.Cor().String())
	require.True(t, v.Ativo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDDeveRetornarVeiculoNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(gorm.ErrRecordNotFound)

	v, err := repository.BuscarPorID(context.Background(), 999)

	require.Nil(t, v)
	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	v, err := repository.BuscarPorID(context.Background(), 1)

	require.Nil(t, v)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorPlaca(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(veiculoRows())

	placa, err := domain.NewPlaca("ABC1D23")
	require.NoError(t, err)

	v, err := repository.BuscarPorPlaca(context.Background(), placa)

	require.NoError(t, err)
	require.NotNil(t, v)
	require.Equal(t, "ABC1D23", v.Placa().String())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorPlacaDeveRetornarVeiculoNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(gorm.ErrRecordNotFound)

	placa, err := domain.NewPlaca("ZZZ9Z99")
	require.NoError(t, err)

	v, err := repository.BuscarPorPlaca(context.Background(), placa)

	require.Nil(t, v)
	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorPlacaDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	placa, err := domain.NewPlaca("ABC1D23")
	require.NoError(t, err)

	v, err := repository.BuscarPorPlaca(context.Background(), placa)

	require.Nil(t, v)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizar(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	v := novoVeiculoValido(t)
	v.AtribuirID(1)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Atualizar(context.Background(), v)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarVeiculoNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	v := novoVeiculoValido(t)
	v.AtribuirID(999)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.Atualizar(context.Background(), v)

	require.ErrorIs(t, err, domain.ErrVeiculoNaoEncontrado)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	v := novoVeiculoValido(t)
	v.AtribuirID(1)

	erroBanco := errors.New("erro ao atualizar veiculo")

	mock.ExpectExec("UPDATE .*").WillReturnError(erroBanco)

	err := repository.Atualizar(context.Background(), v)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}
