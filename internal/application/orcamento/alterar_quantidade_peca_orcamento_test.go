package orcamento_test

import (
	"context"
	"testing"
	"time"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	pecamocks "github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	sharedmocks "github.com/muriloperosa/soat-architecture/internal/domain/shared/mocks"
	"github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func orcamentoPersistidoComPeca() *domainorcamento.Orcamento {
	agora := time.Now()
	return domainorcamento.ReidratarOrcamento(
		100,
		42,
		nil,
		[]domainorcamento.ItemPeca{
			domainorcamento.ReidratarItemPeca(11, 100, 7, "Pastilha", 2, 50),
		},
		0,
		100,
		100,
		"",
		2,
		agora,
		agora,
	)
}

func TestAlterarQuantidadePecaOrcamentoUseCase_AprovadoRemoveReservaEReenvia(t *testing.T) {
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)
	runner := &helpers.TransactionRunnerMock{}

	os := osAguardandoAprovacao(t, 10)
	require.NoError(t, os.AprovarOrcamento())
	orcamento := orcamentoPersistidoComPeca()
	reserva, err := domainreservapeca.NewReservaPeca(42, 7, 2)
	require.NoError(t, err)
	reserva.AtribuirID(9)
	peca, err := domainpeca.NewPeca("Pastilha", "Bosch", "Pastilha dianteira", 50, 10, 2, 2)
	require.NoError(t, err)
	peca.AtribuirID(7)

	cliente, err := domaincliente.NewCliente(
		"52998224725",
		domaincliente.TipoPessoaFisica,
		"Cliente Teste",
		"cliente@teste.com",
		"44999991234",
		"senha123",
		2,
	)
	require.NoError(t, err)
	cliente.DefinirID(10)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamento, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return([]*domainreservapeca.ReservaPeca{reserva}, nil)
	pecaRepo.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(7)).Return(peca, nil)
	reservaRepo.EXPECT().Remover(mock.Anything, uint64(42), uint64(7)).Return(nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, orcamento).Return(nil)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(&cliente, nil)
	emailSender.EXPECT().Enviar(mock.Anything, "cliente@teste.com", mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(nil)

	uc := app.NewAlterarQuantidadePecaOrcamentoUseCase(
		orcamentoRepo,
		osRepo,
		reservaRepo,
		pecaRepo,
		clienteRepo,
		runner,
		emailSender,
	)

	out, err := uc.Executar(context.Background(), app.AlterarQuantidadePecaOrcamentoInput{
		OrdemServicoID: 42,
		ItemPecaID:     11,
		Quantidade:     4,
		UsuarioID:      3,
	})

	require.NoError(t, err)
	require.Equal(t, 4, out.ItensPeca[0].Quantidade)
	require.Equal(t, 200.0, out.ValorItemPecas)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
	require.Equal(t, 1, runner.Calls)
}

func TestAlterarQuantidadePecaOrcamentoUseCase_EmDiagnosticoApenasEdita(t *testing.T) {
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	runner := &helpers.TransactionRunnerMock{}

	os, err := domainordemservico.NewOrdemServico("OS-20260830-a1b2c3d4e5f6", 10, 20, 1000, "", "", 1)
	require.NoError(t, err)
	os.AtribuirID(42)
	require.NoError(t, os.IniciarDiagnostico(2))
	orcamento := orcamentoPersistidoComPeca()

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamento, nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, orcamento).Return(nil)

	uc := app.NewAlterarQuantidadePecaOrcamentoUseCase(
		orcamentoRepo,
		osRepo,
		nil,
		nil,
		nil,
		runner,
		nil,
	)

	out, err := uc.Executar(context.Background(), app.AlterarQuantidadePecaOrcamentoInput{
		OrdemServicoID: 42,
		ItemPecaID:     11,
		Quantidade:     3,
		UsuarioID:      2,
	})

	require.NoError(t, err)
	require.Equal(t, 3, out.ItensPeca[0].Quantidade)
	require.Equal(t, domainordemservico.StatusEmDiagnostico, os.Status())
}
