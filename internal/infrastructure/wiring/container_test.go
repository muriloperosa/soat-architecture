package wiring

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
	gormmysql "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestNewContainerSemConfiguracao(t *testing.T) {
	container, err := NewContainer(nil, nil)
	require.NoError(t, err)

	require.NotNil(t, container)
	require.Nil(t, container.JWTAuth)
	require.Nil(t, container.LoginInternoUC)
	require.NotNil(t, container.RefreshTokensRepo)
	require.NotNil(t, container.UsuarioRepo)
	require.NotNil(t, container.ClienteRepository)
	require.NotNil(t, container.PecaRepo)
	require.NotNil(t, container.ReservaPecaRepo)
	require.NotNil(t, container.TransactionRunner)
	require.NotNil(t, container.EmailSender)
	require.NotNil(t, container.VeiculoRepo)
	require.NotNil(t, container.ServicoRepo)
	require.NotNil(t, container.OrdemServicoRepo)
	require.NotNil(t, container.OrcamentoRepo)
	require.NotNil(t, container.RelatorioRepo)
}

func TestNewContainerComConfiguracao(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	t.Cleanup(func() { _ = sqlDB.Close() })

	db, err := gorm.Open(gormmysql.New(gormmysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	container, err := NewContainer(&config.Config{
		JWTSecret:                "test-secret",
		JWTAccessTokenTTLMinutes: 15,
		JWTRefreshTokenTTLHours:  24,
	}, db)
	require.NoError(t, err)

	require.NotNil(t, container.JWTAuth)
	require.NotNil(t, container.LoginInternoUC)
	require.NotNil(t, container.LoginClienteUC)
	require.NotNil(t, container.RefreshUC)
	require.NotNil(t, container.LogoutUC)
	require.NotNil(t, container.CriarUsuarioUC)
	require.NotNil(t, container.BuscarUsuarioLogadoUC)
	require.NotNil(t, container.CriarClienteUseCase)
	require.NotNil(t, container.ListarClientesUseCase)
	require.NotNil(t, container.CadastrarPecaUC)
	require.NotNil(t, container.ConsultarDisponibilidadeUC)
	require.NotNil(t, container.CadastrarVeiculoUC)
	require.NotNil(t, container.ListarVeiculosUC)
	require.NotNil(t, container.CriarServicoUC)
	require.NotNil(t, container.InativarServicoUC)
	require.NotNil(t, container.AbrirOrdemServicoUC)
	require.NotNil(t, container.ListarOrdensServicoUC)
	require.NotNil(t, container.GerarOrcamentoUC)
	require.NotNil(t, container.RejeitarOrcamentoUC)
	require.NotNil(t, container.ConsultarTransicaoStatusUC)
	require.NotNil(t, container.UsuarioStatusRepo)
	require.NotNil(t, container.ClienteStatusRepo)
}
