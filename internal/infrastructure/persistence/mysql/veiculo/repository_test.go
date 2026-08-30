package veiculo

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
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

func TestRepositoryListarRetornaPagina(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `veiculos`").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery("SELECT .* FROM `veiculos` ORDER BY id ASC LIMIT \\?").
		WithArgs(20).
		WillReturnRows(veiculoRows())

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.NoError(t, err)

	require.Len(t, page.Items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, "id", page.Order)
	require.Equal(t, domainquery.DirectionASC, page.Direction)

	v := page.Items[0]

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

func TestRepositoryListarVazioRetornaPaginaVazia(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `veiculos`").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(0),
		)

	mock.
		ExpectQuery("SELECT .* FROM `veiculos` ORDER BY id ASC LIMIT \\?").
		WithArgs(20).
		WillReturnRows(
			sqlmock.NewRows([]string{
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
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, "id", page.Order)
	require.Equal(t, domainquery.DirectionASC, page.Direction)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarAplicaPaginacaoEOrdenacao(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	params := domainquery.Params{
		Page:      2,
		Order:     "ano",
		Direction: domainquery.DirectionDESC,
	}

	mock.
		ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `veiculos`")).
		WillReturnRows(
			sqlmock.NewRows([]string{"count(*)"}).
				AddRow(50),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM `veiculos` ORDER BY ano DESC LIMIT \\? OFFSET \\?",
		).
		WithArgs(20, 20).
		WillReturnRows(veiculoRows())

	page, err := repository.Listar(
		context.Background(),
		params,
	)

	require.NoError(t, err)

	require.Len(t, page.Items, 1)
	require.Equal(t, int64(50), page.Total)
	require.Equal(t, 2, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, 3, page.TotalPages)
	require.Equal(t, "ano", page.Order)
	require.Equal(t, domainquery.DirectionDESC, page.Direction)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarAplicaFiltroPorMarca(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	params := domainquery.Params{
		Filters: []domainquery.Filter{
			{
				Field:    "marca",
				Operator: domainquery.OperatorLike,
				Value:    "Fiat",
			},
		},
	}

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM `veiculos` WHERE marca LIKE \\?",
		).
		WithArgs("%Fiat%").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM `veiculos` WHERE marca LIKE \\? ORDER BY id ASC LIMIT \\?",
		).
		WithArgs("%Fiat%", 20).
		WillReturnRows(veiculoRows())

	page, err := repository.Listar(
		context.Background(),
		params,
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, int64(1), page.Total)

	require.Equal(t, "Fiat", page.Items[0].Marca())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarAplicaFiltroPorAno(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	params := domainquery.Params{
		Filters: []domainquery.Filter{
			{
				Field:    "ano",
				Operator: domainquery.OperatorEqual,
				Value:    "2020",
			},
		},
	}

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM `veiculos` WHERE ano = \\?",
		).
		WithArgs(uint64(2020)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM `veiculos` WHERE ano = \\? ORDER BY id ASC LIMIT \\?",
		).
		WithArgs(uint64(2020), 20).
		WillReturnRows(veiculoRows())

	page, err := repository.Listar(
		context.Background(),
		params,
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, uint16(2020), page.Items[0].Ano())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarAplicaFiltroPorQuilometragemAtual(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	params := domainquery.Params{
		Filters: []domainquery.Filter{
			{
				Field:    "quilometragem_atual",
				Value:    "15000",
				Operator: domainquery.OperatorEqual,
			},
		},
	}

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM `veiculos` WHERE quilometragem_atual = \\?",
		).
		WithArgs(uint64(15000)).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM `veiculos` WHERE quilometragem_atual = \\? ORDER BY id ASC LIMIT \\?",
		).
		WithArgs(uint64(15000), 20).
		WillReturnRows(veiculoRows())

	page, err := repository.Listar(
		context.Background(),
		params,
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)

	require.Equal(
		t,
		uint32(15000),
		page.Items[0].QuilometragemAtual(),
	)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarAplicaFiltroPorAtivo(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	params := domainquery.Params{
		Filters: []domainquery.Filter{
			{
				Field:    "ativo",
				Value:    "true",
				Operator: domainquery.OperatorEqual,
			},
		},
	}

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM `veiculos` WHERE ativo = \\?",
		).
		WithArgs(true).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery(
			"SELECT .* FROM `veiculos` WHERE ativo = \\? ORDER BY id ASC LIMIT \\?",
		).
		WithArgs(true, 20).
		WillReturnRows(veiculoRows())

	page, err := repository.Listar(
		context.Background(),
		params,
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.True(t, page.Items[0].Ativo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarDeveRetornarErroNaContagem(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao contar veiculos")

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `veiculos`").
		WillReturnError(erroBanco)

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.ErrorIs(t, err, erroBanco)
	require.Empty(t, page.Items)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarDeveRetornarErroAoBuscarRegistros(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	erroBanco := errors.New("erro ao listar veiculos")

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `veiculos`").
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(1),
		)

	mock.
		ExpectQuery("SELECT .* FROM `veiculos` ORDER BY id ASC LIMIT \\?").
		WithArgs(20).
		WillReturnError(erroBanco)

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.ErrorIs(t, err, erroBanco)
	require.Empty(t, page.Items)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryListarLimitMaiorQueMaximoRetornaErro(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	page, err := repository.Listar(
		context.Background(),
		domainquery.Params{
			Page: -1,
		},
	)

	require.Error(t, err)
	require.Empty(t, page.Items)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizar(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	v := novoVeiculoValido(t)
	v.AtribuirID(1)

	mock.
		ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(0, 1))

	err := repository.Atualizar(context.Background(), v)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarDeveRetornarVeiculoNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewRepository(db)

	v := novoVeiculoValido(t)
	v.AtribuirID(999)

	mock.
		ExpectExec("UPDATE .*").
		WillReturnResult(sqlmock.NewResult(0, 0))

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

	mock.
		ExpectExec("UPDATE .*").
		WillReturnError(erroBanco)

	err := repository.Atualizar(context.Background(), v)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}
