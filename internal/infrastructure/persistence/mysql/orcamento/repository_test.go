package orcamento

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
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

func novoOrcamento(t *testing.T) *domain.Orcamento {
	t.Helper()
	o, err := domain.NewOrcamento(42, "obs", 1)
	require.NoError(t, err)
	return o
}

func orcamentoPersistido(t *testing.T) *domain.Orcamento {
	t.Helper()
	agora := time.Now()
	itensServico := []domain.ItemServico{
		domain.ReidratarItemServico(1, 100, 5, 2, 100.0, shared.RestaurarDuracaoEstimada(60)),
	}
	itensPeca := []domain.ItemPeca{
		domain.ReidratarItemPeca(2, 100, 7, "Pastilha", 3, 50.0),
	}
	return domain.ReidratarOrcamento(100, 42, itensServico, itensPeca, 200.0, 150.0, 350.0, "obs", 1, agora, agora)
}

func TestRepositorySalvarPersisteOrcamentoESeusItensNaMesmaTransacao(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)
	o := novoOrcamento(t)
	require.NoError(t, o.AdicionarItemServico(5, 2, 100.0, 60))
	require.NoError(t, o.AdicionarItemPeca(7, "Pastilha", 3, 50.0))

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `orcamentos`").WillReturnResult(sqlmock.NewResult(100, 1))
	mock.ExpectExec("INSERT INTO `orcamentos_servicos`").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO `orcamentos_pecas`").WillReturnResult(sqlmock.NewResult(2, 1))
	mock.ExpectCommit()

	err := repository.Salvar(context.Background(), o)

	require.NoError(t, err)
	require.Equal(t, uint64(100), o.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositorySalvarFazRollbackQuandoInsertDeItemFalha(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)
	o := novoOrcamento(t)
	require.NoError(t, o.AdicionarItemServico(5, 2, 100.0, 60))
	erroBanco := errors.New("erro ao inserir item")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO `orcamentos`").WillReturnResult(sqlmock.NewResult(100, 1))
	mock.ExpectExec("INSERT INTO `orcamentos_servicos`").WillReturnError(erroBanco)
	mock.ExpectRollback()

	err := repository.Salvar(context.Background(), o)

	require.ErrorIs(t, err, erroBanco)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemServicoIDComSucesso(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)

	mock.ExpectQuery("SELECT \\* FROM .*orcamentos.*WHERE ordem_servico_id = \\?").
		WithArgs(uint64(42), 1).
		WillReturnRows(orcamentoRows())
	mock.ExpectQuery("SELECT \\* FROM .*orcamentos_pecas.*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "orcamento_id", "peca_id", "descricao", "quantidade", "valor"}))
	mock.ExpectQuery("SELECT \\* FROM .*orcamentos_servicos.*").
		WillReturnRows(sqlmock.NewRows([]string{"id", "orcamento_id", "servico_id", "quantidade", "valor", "tempo_estimado_minutos"}))

	o, err := repository.BuscarPorOrdemServicoID(context.Background(), 42)

	require.NoError(t, err)
	require.Equal(t, uint64(100), o.ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdemServicoIDNaoEncontrado(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)

	mock.ExpectQuery("SELECT \\* FROM .*orcamentos.*WHERE ordem_servico_id = \\?").
		WillReturnError(gorm.ErrRecordNotFound)

	o, err := repository.BuscarPorOrdemServicoID(context.Background(), 999)

	require.Nil(t, o)
	require.ErrorIs(t, err, domain.ErrOrcamentoNaoEncontrado)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdensServicoIDsListaVaziaNaoConsulta(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)

	orcamentos, err := repository.BuscarPorOrdensServicoIDs(context.Background(), []uint64{})

	require.NoError(t, err)
	require.Empty(t, orcamentos)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryBuscarPorOrdensServicoIDsComSucesso(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)

	mock.ExpectQuery("SELECT \\* FROM .*orcamentos.*WHERE ordem_servico_id IN \\(\\?\\)").
		WithArgs(uint64(42)).
		WillReturnRows(orcamentoRows())

	orcamentos, err := repository.BuscarPorOrdensServicoIDs(context.Background(), []uint64{42})

	require.NoError(t, err)
	require.Len(t, orcamentos, 1)
	require.Equal(t, uint64(100), orcamentos[0].ID())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarPersisteCamposEscalaresESincronizaItens(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)
	o := orcamentoPersistido(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `orcamentos`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT `id` FROM .*orcamentos_servicos.*WHERE orcamento_id = \\?").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectQuery("SELECT `id` FROM .*orcamentos_pecas.*WHERE orcamento_id = \\?").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectExec("UPDATE `orcamentos_pecas`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	err := repository.Atualizar(context.Background(), o)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarRetornaNaoEncontradoQuandoNenhumaLinhaAfetada(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)
	o := orcamentoPersistido(t)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `orcamentos`").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectRollback()

	err := repository.Atualizar(context.Background(), o)

	require.ErrorIs(t, err, domain.ErrOrcamentoNaoEncontrado)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestRepositoryAtualizarInsereNovosERemoveAusentesNaSincronizacao(t *testing.T) {
	db, mock := newRepositoryTestDB(t)
	repository := NewOrcamentoRepository(db)
	o := orcamentoPersistido(t)
	require.NoError(t, o.RemoverItemServico(1))
	require.NoError(t, o.AdicionarItemPeca(9, "Amortecedor", 1, 200.0))

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `orcamentos`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT `id` FROM .*orcamentos_servicos.*WHERE orcamento_id = \\?").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectExec("DELETE FROM `orcamentos_servicos`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery("SELECT `id` FROM .*orcamentos_pecas.*WHERE orcamento_id = \\?").
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectExec("UPDATE `orcamentos_pecas`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("INSERT INTO `orcamentos_pecas`").WillReturnResult(sqlmock.NewResult(3, 1))
	mock.ExpectCommit()

	err := repository.Atualizar(context.Background(), o)

	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}

func orcamentoRows() *sqlmock.Rows {
	agora := time.Now()
	return sqlmock.NewRows([]string{"id", "ordem_servico_id", "valor_item_servicos", "valor_item_pecas", "valor_total", "observacoes", "criado_por", "criado_em", "atualizado_em"}).
		AddRow(100, 42, 200.0, 150.0, 350.0, "obs", 1, agora, agora)
}
