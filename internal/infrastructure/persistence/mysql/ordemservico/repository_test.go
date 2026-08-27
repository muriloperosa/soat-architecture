package ordemservico

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
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
	require.NoError(t, os.IniciarDiagnostico(7, "Diagnóstico iniciado"))

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
	require.NoError(t, os.IniciarDiagnostico(7, "Diagnóstico iniciado"))
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
