package peca

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
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

	mock.
		ExpectExec("INSERT INTO .*").
		WillReturnResult(sqlmock.NewResult(1, 1))

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

	mock.
		ExpectExec("INSERT INTO .*").
		WillReturnError(erroBanco)

	err := repository.Salvar(context.Background(), p)

	require.ErrorIs(t, err, erroBanco)
	require.Equal(t, uint64(0), p.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorID(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT .* FROM .*").
		WillReturnRows(pecaRows())

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

	mock.
		ExpectQuery("SELECT .* FROM .*").
		WillReturnError(gorm.ErrRecordNotFound)

	p, err := repository.BuscarPorID(context.Background(), 999)

	require.Nil(t, p)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.
		ExpectQuery("SELECT .* FROM .*").
		WillReturnError(erroBanco)

	p, err := repository.BuscarPorID(context.Background(), 1)

	require.Nil(t, p)
	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDComBloqueio(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT .* FROM .* FOR UPDATE").
		WillReturnRows(pecaRows())

	p, err := repository.BuscarPorIDComBloqueio(
		context.Background(),
		1,
	)

	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, uint64(1), p.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDComBloqueioDeveRetornarPecaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT .* FROM .* FOR UPDATE").
		WillReturnError(gorm.ErrRecordNotFound)

	p, err := repository.BuscarPorIDComBloqueio(
		context.Background(),
		999,
	)

	require.Nil(t, p)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDComBloqueioDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.
		ExpectQuery("SELECT .* FROM .* FOR UPDATE").
		WillReturnError(erroBanco)

	p, err := repository.BuscarPorIDComBloqueio(
		context.Background(),
		1,
	)

	require.Nil(t, p)
	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDComBloqueio_ParticipaDaTransacaoDoContexto(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	runner := mysql.NewTransactionRunner(db)

	mock.ExpectBegin()

	mock.
		ExpectQuery("SELECT .* FROM .* FOR UPDATE").
		WillReturnRows(pecaRows())

	mock.ExpectCommit()

	err := runner.Executar(
		context.Background(),
		func(ctx context.Context) error {
			p, err := repository.BuscarPorIDComBloqueio(ctx, 1)

			require.NoError(t, err)
			require.NotNil(t, p)

			return nil
		},
	)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorCodigo(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT .* FROM .*").
		WillReturnRows(pecaRows())

	p, err := repository.BuscarPorCodigo(
		context.Background(),
		"PC123",
	)

	require.NoError(t, err)
	require.NotNil(t, p)
	require.Equal(t, "PC123", p.Codigo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorCodigoDeveRetornarPecaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT .* FROM .*").
		WillReturnError(gorm.ErrRecordNotFound)

	p, err := repository.BuscarPorCodigo(
		context.Background(),
		"INEXISTENTE",
	)

	require.Nil(t, p)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorCodigoDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao consultar banco")

	mock.
		ExpectQuery("SELECT .* FROM .*").
		WillReturnError(erroBanco)

	p, err := repository.BuscarPorCodigo(
		context.Background(),
		"PC123",
	)

	require.Nil(t, p)
	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_RetornaPagina(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM .*").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery("SELECT .* FROM .* ORDER BY .*id.* ASC LIMIT .*").
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.NoError(t, err)

	require.Len(t, page.Items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, 0, page.Offset)
	require.Equal(t, 20, page.Limit)
	require.Equal(t, "id", page.Order)
	require.Equal(t, domainquery.DirectionASC, page.Direction)

	require.Equal(t, uint64(1), page.Items[0].ID())
	require.Equal(t, "PC123", page.Items[0].Codigo())
	require.Equal(t, "Peca 1", page.Items[0].Nome())
	require.Equal(t, "Marca 1", page.Items[0].Marca())
	require.Equal(t, 100.0, page.Items[0].Preco())
	require.Equal(t, 10, page.Items[0].QuantidadeEmEstoque())
	require.Equal(t, 5, page.Items[0].EstoqueMinimo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_VazioRetornaPaginaVazia(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM .*").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	mock.
		ExpectQuery("SELECT .* FROM .* ORDER BY .*id.* ASC LIMIT .*").
		WillReturnRows(
			sqlmock.NewRows([]string{
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
			}),
		)

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.NoError(t, err)
	require.NotNil(t, page.Items)
	require.Empty(t, page.Items)
	require.Equal(t, int64(0), page.Total)
	require.Equal(t, 20, page.Limit)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_AplicaPaginacaoEOrdenacao(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM .*").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(10),
		)

	mock.
		ExpectQuery("SELECT .* FROM .* ORDER BY .*preco.* DESC LIMIT .* OFFSET .*").
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Offset:    5,
			Limit:     2,
			Order:     "preco",
			Direction: domainquery.DirectionDESC,
		},
	)

	require.NoError(t, err)

	require.Equal(t, int64(10), page.Total)
	require.Equal(t, 5, page.Offset)
	require.Equal(t, 2, page.Limit)
	require.Equal(t, "preco", page.Order)
	require.Equal(t, domainquery.DirectionDESC, page.Direction)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_AplicaFiltroPorNome(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM .* WHERE .*nome.* LIKE .*",
		).
		WithArgs("%Peca%").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM .* WHERE .*nome.* LIKE .* ORDER BY .*id.* ASC LIMIT .*",
		).
		WithArgs("%Peca%", 20).
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Filters: []domainquery.Filter{
				{
					Field:    "nome",
					Operator: domainquery.OperatorLike,
					Value:    "Peca",
				},
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_AplicaFiltroPorMarca(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM .* WHERE .*marca.* LIKE .*",
		).
		WithArgs("%Marca 1%").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM .* WHERE .*marca.* LIKE .* ORDER BY .*id.* ASC LIMIT .*",
		).
		WithArgs("%Marca 1%", 20).
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Filters: []domainquery.Filter{
				{
					Field:    "marca",
					Operator: domainquery.OperatorLike,
					Value:    "Marca 1",
				},
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_AplicaFiltroPorPreco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM .* WHERE .*preco.* = .*",
		).
		WithArgs(100.0).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM .* WHERE .*preco.* = .* ORDER BY .*id.* ASC LIMIT .*",
		).
		WithArgs(100.0, 20).
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Filters: []domainquery.Filter{
				{
					Field:    "preco",
					Operator: domainquery.OperatorEqual,
					Value:    "100",
				},
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_AplicaFiltroPorQuantidadeEmEstoque(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM .* WHERE .*quantidade_em_estoque.* = .*",
		).
		WithArgs(uint64(10)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM .* WHERE .*quantidade_em_estoque.* = .* ORDER BY .*id.* ASC LIMIT .*",
		).
		WithArgs(uint64(10), 20).
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Filters: []domainquery.Filter{
				{
					Field:    "quantidade_em_estoque",
					Operator: domainquery.OperatorEqual,
					Value:    "10",
				},
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_AplicaFiltroPorEstoqueMinimo(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM .* WHERE .*estoque_minimo.* = .*",
		).
		WithArgs(uint64(5)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM .* WHERE .*estoque_minimo.* = .* ORDER BY .*id.* ASC LIMIT .*",
		).
		WithArgs(uint64(5), 20).
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Filters: []domainquery.Filter{
				{
					Field:    "estoque_minimo",
					Operator: domainquery.OperatorEqual,
					Value:    "5",
				},
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_AplicaFiltroPorAtivo(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM .* WHERE .*ativo.* = .*",
		).
		WithArgs(true).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM .* WHERE .*ativo.* = .* ORDER BY .*id.* ASC LIMIT .*",
		).
		WithArgs(true, 20).
		WillReturnRows(pecaRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Filters: []domainquery.Filter{
				{
					Field:    "ativo",
					Operator: domainquery.OperatorEqual,
					Value:    "true",
				},
			},
		},
	)

	require.NoError(t, err)
	require.Equal(t, int64(1), page.Total)
	require.Len(t, page.Items, 1)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_DeveRetornarErroNaContagem(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao contar pecas")

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM .*").
		WillReturnError(erroBanco)

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.ErrorIs(t, err, erroBanco)
	require.Empty(t, page.Items)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_DeveRetornarErroAoBuscarRegistros(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao listar pecas")

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM .*").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery("SELECT .* FROM .* ORDER BY .*id.* ASC LIMIT .*").
		WillReturnError(erroBanco)

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.ErrorIs(t, err, erroBanco)
	require.Empty(t, page.Items)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListar_LimitMaiorQueMaximoRetornaErro(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Limit: 101,
		},
	)

	require.Error(t, err)
	require.Empty(t, page.Items)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizar(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)
	p.AtribuirID(1)

	mock.
		ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Atualizar(
		context.Background(),
		p,
	)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarPecaNaoEncontrada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)
	p.AtribuirID(999)

	mock.
		ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(0, 0))

	err := repository.Atualizar(
		context.Background(),
		p,
	)

	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarErroDoBanco(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	p := novaPecaValida(t)
	p.AtribuirID(1)

	erroBanco := errors.New("erro ao atualizar peca")

	mock.
		ExpectExec("UPDATE .*").
		WillReturnError(erroBanco)

	err := repository.Atualizar(
		context.Background(),
		p,
	)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}
