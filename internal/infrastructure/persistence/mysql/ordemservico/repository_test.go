package ordemservico

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
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

func novaOrdemServico(t *testing.T) *domain.OrdemServico {
	t.Helper()
	os, err := domain.NewOrdemServico("OS-20260826-a1b2c3d4e5f6", 10, 20, 52_300, "", "Ruído no motor", 30)
	require.NoError(t, err)
	return os
}

func TestRepositorySalvarPersisteOrdemEHistoricoNaMesmaTransacao(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)
	os := novaOrdemServico(t)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `ordens_servico`").WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec("INSERT INTO `historicos_status`").WillReturnResult(sqlmock.NewResult(7, 1))
	mock.ExpectCommit()

	err := repository.Salvar(context.Background(), os)

	require.NoError(t, err)
	require.Equal(t, uint64(42), os.ID())
	require.Equal(t, uint64(42), os.HistoricoStatus()[0].OrdemServicoID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySalvarFazRollbackQuandoHistoricoFalha(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)
	os := novaOrdemServico(t)
	erroBanco := errors.New("erro ao inserir histórico")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `ordens_servico`").WillReturnResult(sqlmock.NewResult(42, 1))
	mock.ExpectExec("INSERT INTO `historicos_status`").WillReturnError(erroBanco)
	mock.ExpectRollback()

	err := repository.Salvar(context.Background(), os)

	require.ErrorIs(t, err, erroBanco)
	require.Zero(t, os.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarPersisteEstadoENovoHistoricoNaMesmaTransacao(t *testing.T) {
	db, mockDB := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)
	os := ordemServicoPersistida(t)
	require.NoError(t, os.IniciarDiagnostico(7))

	mockDB.ExpectBegin()
	mockDB.ExpectExec("UPDATE `ordens_servico`").WillReturnResult(sqlmock.NewResult(0, 1))
	mockDB.ExpectExec("INSERT INTO `historicos_status`").WillReturnResult(sqlmock.NewResult(8, 1))
	mockDB.ExpectCommit()

	err := repository.Atualizar(context.Background(), os)

	require.NoError(t, err)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestRepositoryAtualizarFazRollbackQuandoNovoHistoricoFalha(t *testing.T) {
	db, mockDB := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)
	os := ordemServicoPersistida(t)
	require.NoError(t, os.IniciarDiagnostico(7))
	erroBanco := errors.New("erro ao inserir novo histórico")

	mockDB.ExpectBegin()
	mockDB.ExpectExec("UPDATE `ordens_servico`").WillReturnResult(sqlmock.NewResult(0, 1))
	mockDB.ExpectExec("INSERT INTO `historicos_status`").WillReturnError(erroBanco)
	mockDB.ExpectRollback()

	err := repository.Atualizar(context.Background(), os)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func ordemServicoPersistida(t *testing.T) *domain.OrdemServico {
	t.Helper()
	numero, err := domain.NewNumeroOrdemServico("OS-20260827-a1b2c3d4e5f6")
	require.NoError(t, err)
	agora := time.Now()
	historico := domain.ReidratarHistoricoStatus(7, 42, domain.StatusRecebida, agora, 3, "")

	return domain.ReidratarOrdemServico(
		42,
		numero,
		10,
		20,
		52_300,
		domain.StatusRecebida,
		"",
		"Ruído no motor",
		3,
		[]domain.HistoricoStatus{historico},
		agora,
		agora,
	)
}

func TestRepositoryListarComPaginacaoOrdenacaoEFiltros(t *testing.T) {
	db, mockDB := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)

	params := query.Params{
		Page:      2,
		Order:     "data_cadastro",
		Direction: query.DirectionDESC,
		Filters: []query.Filter{
			{
				Field:    "status",
				Operator: query.OperatorEqual,
				Value:    "RECEBIDA",
			},
			{
				Field:    "cliente_id",
				Operator: query.OperatorEqual,
				Value:    "10",
			},
		},
	}

	mockDB.
		ExpectQuery(
			"SELECT count\\(\\*\\) FROM .*ordens_servico.*WHERE.*status.*=.*\\?.*AND.*cliente_id.*=.*\\?",
		).
		WithArgs(
			"RECEBIDA",
			uint64(10),
		).
		WillReturnRows(
			sqlmock.NewRows([]string{"count"}).
				AddRow(21),
		)

	mockDB.
		ExpectQuery(
			"SELECT \\* FROM .*ordens_servico.*WHERE.*status.*=.*\\?.*AND.*cliente_id.*=.*\\?.*ORDER BY.*data_cadastro.*DESC.*LIMIT.*OFFSET.*",
		).
		WithArgs(
			"RECEBIDA",
			uint64(10),
			20,
			20,
		).
		WillReturnRows(ordemServicoRows())

	page, err := repository.Listar(context.Background(), params)

	require.NoError(t, err)
	require.Len(t, page.Items, 1)
	require.Equal(t, int64(21), page.Total)
	require.Equal(t, 2, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Equal(t, 2, page.TotalPages)
	require.Equal(t, "data_cadastro", page.Order)
	require.Equal(t, query.DirectionDESC, page.Direction)
	require.Equal(t, uint64(42), page.Items[0].ID())
	require.Empty(t, page.Items[0].HistoricoStatus())

	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestRepositoryListarUsaPadroes(t *testing.T) {
	db, mockDB := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)

	mockDB.ExpectQuery("SELECT count\\(\\*\\) FROM .*ordens_servico.*").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(0))
	mockDB.ExpectQuery("SELECT \\* FROM .*ordens_servico.* ORDER BY id ASC LIMIT .*").
		WillReturnRows(sqlmock.NewRows(ordemServicoColumns()))

	page, err := repository.Listar(context.Background(), query.Params{})

	require.NoError(t, err)
	require.Equal(t, 1, page.Page)
	require.Equal(t, 20, page.PageSize)
	require.Zero(t, page.TotalPages)
	require.Equal(t, "id", page.Order)
	require.Equal(t, query.DirectionASC, page.Direction)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDComBloqueio(t *testing.T) {
	db, mockDB := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)
	agora := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)

	mockDB.ExpectQuery("SELECT .* FROM .*ordens_servico.* FOR UPDATE").
		WillReturnRows(ordemServicoRows())
	mockDB.ExpectQuery("SELECT .* FROM .*historicos_status.*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "ordem_servico_id", "status", "alterado_por", "motivo", "alterado_em"}).
			AddRow(1, 42, "RECEBIDA", 30, "", agora))

	os, err := repository.BuscarPorIDComBloqueio(context.Background(), 42)

	require.NoError(t, err)
	require.Equal(t, uint64(42), os.ID())
	require.Equal(t, domain.StatusRecebida, os.Status())
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func TestRepositoryBuscarPorIDComBloqueioDeveRetornarNaoEncontrada(t *testing.T) {
	db, mockDB := newRepositoryTestDB(t)
	repository := NewOrdemServicoRepository(db)

	mockDB.ExpectQuery("SELECT .* FROM .*ordens_servico.* FOR UPDATE").
		WillReturnError(gorm.ErrRecordNotFound)

	os, err := repository.BuscarPorIDComBloqueio(context.Background(), 999)

	require.Nil(t, os)
	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
	require.NoError(t, mockDB.ExpectationsWereMet())
}

func ordemServicoColumns() []string {
	return []string{"id", "numero", "cliente_id", "veiculo_id", "quilometragem_entrada", "status", "diagnostico", "observacoes", "criado_por", "data_cadastro", "data_atualizacao"}
}

func ordemServicoRows() *sqlmock.Rows {
	agora := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	return sqlmock.NewRows(ordemServicoColumns()).AddRow(
		42,
		"OS-20260830-a1b2c3d4e5f6",
		10,
		20,
		52300,
		"RECEBIDA",
		"",
		"Ruído no motor",
		30,
		agora,
		agora,
	)
}
