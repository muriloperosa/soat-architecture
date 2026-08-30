package servico_test

import (
	"context"
	"errors"
	"regexp"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/servico"
	mysqlservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/servico"
	test_helpers "github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

func novoServico(t *testing.T) *domainservico.Servico {
	t.Helper()
	s, err := domainservico.NewServico("Troca de óleo", "Troca de óleo e filtro", 150.50, 60, 1)
	require.NoError(t, err)
	return s
}

func colunasServico() []string {
	return []string{"id", "nome", "descricao", "preco_base", "tempo_estimado_minutos", "criado_por", "ativo", "data_cadastro", "data_atualizacao"}
}

func TestServicoRepository_Salvar_ExecutaInsertEPreencheID(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `servicos`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectCommit()

	s := novoServico(t)
	err := repo.Salvar(context.Background(), s)
	require.NoError(t, err)
	require.Equal(t, uint64(1), s.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Salvar_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `servicos`").WillReturnError(errors.New("conexao recusada"))
	mock.ExpectRollback()

	s := novoServico(t)
	err := repo.Salvar(context.Background(), s)
	require.Error(t, err)
	require.Zero(t, s.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_BuscarPorID_NaoEncontrado_RetornaErrServicoNaoEncontrado(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `servicos`").WillReturnError(gorm.ErrRecordNotFound)

	s, err := repo.BuscarPorID(context.Background(), 99)
	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
	require.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_BuscarPorID_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `servicos`").WillReturnError(errors.New("conexao recusada"))

	s, err := repo.BuscarPorID(context.Background(), 1)
	require.Error(t, err)
	require.False(t, errors.Is(err, domainservico.ErrServicoNaoEncontrado))
	require.Nil(t, s)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_BuscarPorID_RetornaEntidade(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	agora := time.Now()

	rows := sqlmock.
		NewRows(colunasServico()).
		AddRow(1, "Troca de óleo", "Troca de óleo e filtro", 150.50, 60, 9, true, agora, agora)

	mock.ExpectQuery("SELECT \\* FROM `servicos`").WillReturnRows(rows)

	s, err := repo.BuscarPorID(context.Background(), 1)

	require.NoError(t, err)
	require.Equal(t, uint64(1), s.ID())
	require.Equal(t, "Troca de óleo", s.Nome())
	require.Equal(t, 150.50, s.PrecoBase())
	require.Equal(t, 60, s.TempoEstimado().Minutos())
	require.Equal(t, uint64(9), s.CriadoPor())
	require.True(t, s.Ativo())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_RetornaPagina(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	agora := time.Now()

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `servicos`").
		WillReturnRows(
			sqlmock.NewRows([]string{"count(*)"}).
				AddRow(2),
		)

	rows := sqlmock.
		NewRows(colunasServico()).
		AddRow(
			1,
			"Troca de óleo",
			"descrição",
			150.50,
			60,
			1,
			true,
			agora,
			agora,
		).
		AddRow(
			2,
			"Alinhamento",
			"alinhamento",
			200.00,
			90,
			1,
			false,
			agora,
			agora,
		)

	mock.
		ExpectQuery("SELECT \\* FROM `servicos` ORDER BY id ASC LIMIT \\?").
		WithArgs(20).
		WillReturnRows(rows)

	page, err := repo.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.NoError(t, err)

	require.Len(t, page.Items, 2)
	require.Equal(t, int64(2), page.Total)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, "id", page.Order)
	require.Equal(t, domainquery.DirectionASC, page.Direction)

	require.Equal(t, "Troca de óleo", page.Items[0].Nome())
	require.Equal(t, "Alinhamento", page.Items[1].Nome())
	require.False(t, page.Items[1].Ativo())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_Vazio_RetornaPaginaVazia(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `servicos`").
		WillReturnRows(
			sqlmock.NewRows([]string{"count(*)"}).
				AddRow(0),
		)

	rows := sqlmock.NewRows(colunasServico())

	mock.
		ExpectQuery("SELECT \\* FROM `servicos` ORDER BY id ASC LIMIT \\?").
		WithArgs(20).
		WillReturnRows(rows)

	page, err := repo.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.NoError(t, err)

	require.Empty(t, page.Items)
	require.NotNil(t, page.Items)

	require.Equal(t, int64(0), page.Total)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, "id", page.Order)
	require.Equal(t, domainquery.DirectionASC, page.Direction)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_ErroNaContagem_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `servicos`").
		WillReturnError(errors.New("conexao recusada"))

	page, err := repo.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.Error(t, err)
	require.Empty(t, page.Items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_ErroAoBuscarRegistros_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.
		ExpectQuery("SELECT count\\(\\*\\) FROM `servicos`").
		WillReturnRows(
			sqlmock.NewRows([]string{"count(*)"}).
				AddRow(2),
		)

	mock.
		ExpectQuery("SELECT \\* FROM `servicos` ORDER BY id ASC LIMIT \\?").
		WithArgs(20).
		WillReturnError(errors.New("conexao recusada"))

	page, err := repo.Listar(
		context.Background(),
		domainquery.Params{},
	)

	require.Error(t, err)
	require.Empty(t, page.Items)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_AplicaPaginacaoEOrdenacao(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := servico.NewServicoRepository(gdb)

	params := query.Params{
		Page:      2,
		Order:     "nome",
		Direction: query.DirectionDESC,
	}

	mock.
		ExpectQuery(regexp.QuoteMeta("SELECT count(*) FROM `servicos`")).
		WillReturnRows(
			sqlmock.NewRows([]string{"count(*)"}).
				AddRow(50),
		)

	rows := sqlmock.NewRows([]string{
		"id",
		"nome",
		"descricao",
		"preco",
		"duracao_estimada",
		"criado_por",
		"ativo",
		"data_cadastro",
		"data_atualizacao",
	})

	mock.
		ExpectQuery(regexp.QuoteMeta(
			"SELECT * FROM `servicos` ORDER BY nome DESC LIMIT ? OFFSET ?",
		)).
		WithArgs(20, 20).
		WillReturnRows(rows)

	page, err := repo.Listar(context.Background(), params)

	require.NoError(t, err)

	require.Equal(t, int64(50), page.Total)
	require.Equal(t, 2, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, 3, page.TotalPages)
	require.Equal(t, "nome", page.Order)
	require.Equal(t, query.DirectionDESC, page.Direction)

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_AplicaFiltro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	agora := time.Now()

	params := domainquery.Params{
		Filters: []domainquery.Filter{
			{
				Field:    "nome",
				Operator: domainquery.OperatorLike,
				Value:    "óleo",
			},
		},
	}

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM `servicos` WHERE nome LIKE \\?",
		).
		WithArgs("%óleo%").
		WillReturnRows(
			sqlmock.NewRows([]string{"count(*)"}).
				AddRow(1),
		)

	rows := sqlmock.
		NewRows(colunasServico()).
		AddRow(
			1,
			"Troca de óleo",
			"Troca de óleo e filtro",
			150.50,
			60,
			1,
			true,
			agora,
			agora,
		)

	mock.
		ExpectQuery(
			"SELECT \\* FROM `servicos` WHERE nome LIKE \\? ORDER BY id ASC LIMIT \\?",
		).
		WithArgs("%óleo%", 20).
		WillReturnRows(rows)

	page, err := repo.Listar(
		context.Background(),
		params,
	)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, "Troca de óleo", page.Items[0].Nome())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_AplicaFiltroPrecoBase(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	agora := time.Now()

	params := domainquery.Params{
		Filters: []domainquery.Filter{
			{
				Field:    "preco_base",
				Operator: domainquery.OperatorGreaterOrEq,
				Value:    "150.50",
			},
		},
	}

	mock.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM `servicos` WHERE preco_base >= \\?",
		).
		WithArgs(150.50).
		WillReturnRows(
			sqlmock.NewRows([]string{"count(*)"}).
				AddRow(1),
		)

	rows := sqlmock.
		NewRows(colunasServico()).
		AddRow(
			1,
			"Troca de óleo",
			"Troca de óleo e filtro",
			150.50,
			60,
			1,
			true,
			agora,
			agora,
		)

	mock.
		ExpectQuery(
			"SELECT \\* FROM `servicos` WHERE preco_base >= \\? ORDER BY id ASC LIMIT \\?",
		).
		WithArgs(150.50, 20).
		WillReturnRows(rows)

	page, err := repo.Listar(
		context.Background(),
		params,
	)

	require.NoError(t, err)

	require.Len(t, page.Items, 1)
	require.Equal(t, int64(1), page.Total)
	require.Equal(t, 150.50, page.Items[0].PrecoBase())

	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_LimitMaiorQueMaximo_RetornaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	page, err := repo.Listar(
		context.Background(),
		domainquery.Params{
			Page: -1,
		},
	)

	require.Error(t, err)
	require.Empty(t, page.Items)

	// Normalize falha antes de qualquer consulta ao banco.
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Atualizar_ExecutaUpdate(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.ExpectBegin()
	mock.
		ExpectExec("UPDATE `servicos`").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	s := novoServico(t)
	s.AtribuirID(1)

	err := repo.Atualizar(context.Background(), s)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
