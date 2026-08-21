package servico_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
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
	rows := sqlmock.NewRows(colunasServico()).
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

func TestServicoRepository_Listar_RetornaCatalogo(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	agora := time.Now()
	rows := sqlmock.NewRows(colunasServico()).
		AddRow(1, "Troca de óleo", "descrição", 150.50, 60, 1, true, agora, agora).
		AddRow(2, "Alinhamento", "alinhamento", 200.00, 90, 1, false, agora, agora)
	mock.ExpectQuery("SELECT \\* FROM `servicos`").WillReturnRows(rows)

	servicos, err := repo.Listar(context.Background())
	require.NoError(t, err)
	require.Len(t, servicos, 2)
	require.Equal(t, "Troca de óleo", servicos[0].Nome())
	require.Equal(t, "Alinhamento", servicos[1].Nome())
	require.False(t, servicos[1].Ativo())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_Vazio_RetornaSliceVazio(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	rows := sqlmock.NewRows(colunasServico())
	mock.ExpectQuery("SELECT \\* FROM `servicos`").WillReturnRows(rows)

	servicos, err := repo.Listar(context.Background())
	require.NoError(t, err)
	require.Empty(t, servicos)
	require.NotNil(t, servicos)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Listar_ErroDoBanco_PropagaErro(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.ExpectQuery("SELECT \\* FROM `servicos`").WillReturnError(errors.New("conexao recusada"))

	servicos, err := repo.Listar(context.Background())
	require.Error(t, err)
	require.Nil(t, servicos)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestServicoRepository_Atualizar_ExecutaUpdate(t *testing.T) {
	gdb, mock := test_helpers.SetupGormMock(t)
	repo := mysqlservico.NewServicoRepository(gdb)

	mock.ExpectBegin()
	mock.ExpectExec("UPDATE `servicos`").WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	s := novoServico(t)
	s.AtribuirID(1)

	err := repo.Atualizar(context.Background(), s)
	require.NoError(t, err)
	require.NoError(t, mock.ExpectationsWereMet())
}
