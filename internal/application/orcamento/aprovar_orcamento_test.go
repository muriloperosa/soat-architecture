package orcamento_test

import (
	"context"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	pecamocks "github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/test/helpers"
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

func orcamentoComServicoValido(t *testing.T) *domainorcamento.Orcamento {
	t.Helper()
	o, err := domainorcamento.NewOrcamento(42, "", 2)
	require.NoError(t, err)
	o.AtribuirID(100)
	require.NoError(t, o.AdicionarItemServico(1, 1, 100, 60))
	return o
}

func TestAprovarOrcamentoUseCase_ClienteProprietarioAprova(t *testing.T) {
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	o := osAguardandoAprovacao(t, 10)
	orcamento := orcamentoComServicoValido(t)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamento, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return([]*domainreservapeca.ReservaPeca{}, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewAprovarOrcamentoUseCase(osRepo, orcamentoRepo, pecaRepo, reservaRepo, runner)
	out, err := uc.Executar(context.Background(), app.AprovarOrcamentoInput{OrdemServicoID: 42, ClienteID: 10})

	require.NoError(t, err)
	require.Equal(t, "APROVADA", out.Status)
	require.Equal(t, domainordemservico.StatusAprovada, o.Status())
	require.Equal(t, 1, runner.Calls)
	historico := o.HistoricoStatus()
	require.Equal(t, uint64(2), historico[len(historico)-1].AlteradoPor())
}

func TestAprovarOrcamentoUseCase_CriaReservaAPartirDoOrcamento(t *testing.T) {
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	o := osAguardandoAprovacao(t, 10)

	orcamento, err := domainorcamento.NewOrcamento(42, "", 2)
	require.NoError(t, err)
	orcamento.AtribuirID(100)
	require.NoError(t, orcamento.AdicionarItemPeca(7, "Pastilha", 3, 50))

	peca, err := domainpeca.NewPeca("Pastilha", "Bosch", "Pastilha dianteira", 50, 10, 2, 2)
	require.NoError(t, err)
	peca.AtribuirID(7)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamento, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return([]*domainreservapeca.ReservaPeca{}, nil)
	pecaRepo.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(7)).Return(peca, nil)
	reservaRepo.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(7)).Return(2, nil)
	reservaRepo.EXPECT().Salvar(mock.Anything, mock.MatchedBy(func(r *domainreservapeca.ReservaPeca) bool {
		return r.OrdemServicoID() == 42 && r.PecaID() == 7 && r.Quantidade() == 3
	})).Return(nil)
	osRepo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewAprovarOrcamentoUseCase(osRepo, orcamentoRepo, pecaRepo, reservaRepo, runner)
	out, err := uc.Executar(context.Background(), app.AprovarOrcamentoInput{OrdemServicoID: 42, ClienteID: 10})

	require.NoError(t, err)
	require.Equal(t, "APROVADA", out.Status)
	require.Equal(t, 1, runner.Calls)
}

func TestAprovarOrcamentoUseCase_EstoqueInsuficienteNaoAprova(t *testing.T) {
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	o := osAguardandoAprovacao(t, 10)

	orcamento, err := domainorcamento.NewOrcamento(42, "", 2)
	require.NoError(t, err)
	require.NoError(t, orcamento.AdicionarItemPeca(7, "Pastilha", 4, 50))

	peca, err := domainpeca.NewPeca("Pastilha", "Bosch", "Pastilha dianteira", 50, 5, 2, 2)
	require.NoError(t, err)
	peca.AtribuirID(7)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamento, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return([]*domainreservapeca.ReservaPeca{}, nil)
	pecaRepo.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(7)).Return(peca, nil)
	reservaRepo.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(7)).Return(0, nil)

	uc := app.NewAprovarOrcamentoUseCase(osRepo, orcamentoRepo, pecaRepo, reservaRepo, runner)
	_, err = uc.Executar(context.Background(), app.AprovarOrcamentoInput{OrdemServicoID: 42, ClienteID: 10})

	require.ErrorIs(t, err, domainpeca.ErrQuantidadeIndisponivelParaReserva)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, o.Status())
}

func TestAprovarOrcamentoUseCase_ClienteDeOutraOSRetornaForbidden(t *testing.T) {
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	o := osAguardandoAprovacao(t, 10)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)

	uc := app.NewAprovarOrcamentoUseCase(osRepo, orcamentoRepo, pecaRepo, reservaRepo, &helpers.TransactionRunnerMock{})
	_, err := uc.Executar(context.Background(), app.AprovarOrcamentoInput{OrdemServicoID: 42, ClienteID: 99})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindForbidden, appErr.Kind)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, o.Status())
}
