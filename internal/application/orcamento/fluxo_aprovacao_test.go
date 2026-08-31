package orcamento_test

import (
	"context"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	servicomocks "github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func osAguardandoAprovacao(t *testing.T, clienteID uint64) *domainordemservico.OrdemServico {
	t.Helper()
	o, err := domainordemservico.NewOrdemServico("OS-20260830-a1b2c3d4e5f6", clienteID, 20, 1000, "", "", 1)
	require.NoError(t, err)
	o.AtribuirID(42)
	require.NoError(t, o.IniciarDiagnostico(2))
	require.NoError(t, o.InformarDiagnostico("Falha identificada"))
	require.NoError(t, o.EnviarParaAprovacao(2))
	return o
}

func TestAprovarOrcamentoUseCase_ClienteProprietarioAprova(t *testing.T) {
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	o := osAguardandoAprovacao(t, 10)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	repo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewAprovarOrcamentoUseCase(repo)
	out, err := uc.Executar(context.Background(), app.AprovarOrcamentoInput{OrdemServicoID: 42, ClienteID: 10})

	require.NoError(t, err)
	require.Equal(t, "APROVADA", out.Status)
	require.Equal(t, domainordemservico.StatusAprovada, o.Status())
	historico := o.HistoricoStatus()
	require.Equal(t, uint64(2), historico[len(historico)-1].AlteradoPor())
}

func TestAprovarOrcamentoUseCase_ClienteDeOutraOSRetornaForbidden(t *testing.T) {
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	o := osAguardandoAprovacao(t, 10)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)

	uc := app.NewAprovarOrcamentoUseCase(repo)
	_, err := uc.Executar(context.Background(), app.AprovarOrcamentoInput{OrdemServicoID: 42, ClienteID: 99})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindForbidden, appErr.Kind)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, o.Status())
}

func TestRejeitarOrcamentoUseCase_RegistraMotivoNoHistorico(t *testing.T) {
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	o := osAguardandoAprovacao(t, 10)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	repo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewRejeitarOrcamentoUseCase(repo)
	out, err := uc.Executar(context.Background(), app.RejeitarOrcamentoInput{
		OrdemServicoID: 42,
		ClienteID:      10,
		Motivo:         "Valor acima do esperado",
	})

	require.NoError(t, err)
	require.Equal(t, "REJEITADA", out.Status)
	historico := o.HistoricoStatus()
	require.Equal(t, "Valor acima do esperado", historico[len(historico)-1].Motivo())
	require.Equal(t, uint64(2), historico[len(historico)-1].AlteradoPor())
}

func TestAdicionarServicoOrcamentoUseCase_OrcamentoAprovadoEhImutavel(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	o := osAguardandoAprovacao(t, 10)
	require.NoError(t, o.AprovarOrcamento())

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)

	uc := app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, osRepo)
	_, err := uc.Executar(context.Background(), app.AdicionarServicoOrcamentoInput{OrdemServicoID: 42, ServicoID: 1, Quantidade: 1})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoImutavel)
}
