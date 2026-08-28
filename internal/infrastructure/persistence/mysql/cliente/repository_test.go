package cliente

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
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

func novoClienteValido(t *testing.T) *domain.Cliente {
	t.Helper()

	cliente, err := domain.NewCliente(
		"529.982.247-25",
		domain.TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
		1,
	)

	require.NoError(t, err)

	return &cliente
}

func clienteRows() *sqlmock.Rows {
	agora := time.Now()

	return sqlmock.NewRows([]string{
		"id",
		"documento",
		"tipo",
		"nome",
		"email",
		"senha_hash",
		"requer_alterar_senha",
		"criado_por",
		"telefone",
		"ativo",
		"data_cadastro",
		"data_atualizacao",
	}).AddRow(
		uint64(1),
		"52998224725",
		"PF",
		"João Da Silva",
		"joao@email.com",
		"senha123",
		true,
		uint64(7),
		"44999991234",
		true,
		agora,
		agora,
	)
}

func TestRepositoryListarComPaginacaoOrdenacaoEFiltros(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)
	params := query.Params{
		Offset:    1,
		Limit:     1,
		Order:     "nome",
		Direction: query.DirectionDESC,
		Filters: []query.Filter{
			{Field: "nome", Operator: query.OperatorLike, Value: "João"},
			{Field: "ativo", Operator: query.OperatorEqual, Value: "true"},
		},
	}

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*").
		WithArgs("%João%", true).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery("SELECT \\* FROM .* ORDER BY nome DESC LIMIT .* OFFSET .*").
		WillReturnRows(clienteRows())

	page, err := repository.Listar(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, int64(2), page.Total)
	require.Equal(t, 1, page.Offset)
	require.Equal(t, 1, page.Limit)
	require.Equal(t, "nome", page.Order)
	require.Equal(t, query.DirectionDESC, page.Direction)
	require.Equal(t, uint64(7), page.Items[0].CriadoPor())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarUsaPadroes(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT count\\(\\*\\) FROM .*").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mock.ExpectQuery("SELECT \\* FROM .* ORDER BY id ASC LIMIT .*").
		WillReturnRows(clienteRows())

	page, err := repository.Listar(context.Background(), query.Params{})

	require.NoError(t, err)
	require.Equal(t, 20, page.Limit)
	require.Equal(t, "id", page.Order)
	require.Equal(t, query.DirectionASC, page.Direction)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarRejeitaCampoDeFiltroNaoPermitido(t *testing.T) {
	db, _ := newRepositoryTestDB(t)
	repository := NewRepository(db)

	_, err := repository.Listar(context.Background(), query.Params{
		Filters: []query.Filter{{
			Field: "senha_hash", Operator: query.OperatorEqual, Value: "segredo",
		}},
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
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

	cliente := novoClienteValido(t)

	mock.ExpectExec("INSERT INTO .*").WillReturnResult(sqlmock.NewResult(1, 1))

	err := repository.Salvar(context.Background(), cliente)

	require.NoError(t, err)
	require.Equal(t, uint64(1), cliente.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySalvarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	cliente := novoClienteValido(t)

	erroBanco := errors.New("erro ao inserir cliente")

	mock.ExpectExec("INSERT INTO .*").WillReturnError(erroBanco)

	err := repository.Salvar(context.Background(), cliente)

	require.ErrorIs(t, err, erroBanco)
	require.Equal(t, uint64(0), cliente.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorID(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(clienteRows())

	cliente, err := repository.BuscarPorID(context.Background(), 1)

	require.NoError(t, err)
	require.NotNil(t, cliente)

	require.Equal(t, uint64(1), cliente.ID())
	require.Equal(t, "52998224725", cliente.Documento().String())
	require.Equal(t, domain.TipoPessoaFisica, cliente.Tipo())
	require.Equal(t, "João Da Silva", cliente.Nome())
	require.Equal(t, "joao@email.com", cliente.Email().String())
	require.Equal(t, "44999991234", cliente.Telefone().String())
	require.True(t, cliente.Ativo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDDeveRetornarClienteNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(gorm.ErrRecordNotFound)

	cliente, err := repository.BuscarPorID(context.Background(), 999)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	cliente, err := repository.BuscarPorID(context.Background(), 1)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorDocumento(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnRows(clienteRows())

	cliente, err := repository.BuscarPorDocumento(context.Background(), "52998224725")

	require.NoError(t, err)
	require.NotNil(t, cliente)

	require.Equal(t, uint64(1), cliente.ID())
	require.Equal(t, "52998224725", cliente.Documento().String())
	require.Equal(t, domain.TipoPessoaFisica, cliente.Tipo())
	require.Equal(t, "João Da Silva", cliente.Nome())
	require.Equal(t, "joao@email.com", cliente.Email().String())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorDocumentoDeveRetornarClienteNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(gorm.ErrRecordNotFound)

	cliente, err := repository.BuscarPorDocumento(context.Background(), "00000000000")

	require.Nil(t, cliente)
	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorDocumentoDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.ExpectQuery("SELECT .* FROM .*").WillReturnError(erroBanco)

	cliente, err := repository.BuscarPorDocumento(context.Background(), "52998224725")

	require.Nil(t, cliente)
	require.ErrorIs(t, err, erroBanco)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizar(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	cliente := novoClienteValido(t)
	cliente.DefinirID(1)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Atualizar(context.Background(), cliente)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarClienteNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	cliente := novoClienteValido(t)
	cliente.DefinirID(999)

	mock.ExpectExec("UPDATE .*").WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.Atualizar(context.Background(), cliente)

	require.ErrorIs(t, err, domain.ErrClienteNaoEncontrado)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	cliente := novoClienteValido(t)
	cliente.DefinirID(1)

	erroBanco := errors.New("erro ao atualizar cliente")

	mock.ExpectExec("UPDATE .*").WillReturnError(erroBanco)

	err := repository.Atualizar(context.Background(), cliente)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}
