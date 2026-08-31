package peca

import (
	"context"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAlterarQuantidadeReservaPecaUseCase_DesconsideraReservaAtualNaValidacao(t *testing.T) {
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	uc := NewAlterarQuantidadeReservaPecaUseCase(repository, reservaRepository, runner)

	p := pecaComEstoque(t, 10, 2)
	reserva, err := domainreservapeca.NewReservaPeca(10, 1, 3)
	require.NoError(t, err)
	reserva.AtribuirID(5)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(mock.Anything, uint64(10), uint64(1)).Return(reserva, nil).Once()
	// Total 5 = 3 desta OS + 2 de outras OS. Alterar para 6 ainda respeita:
	// 10 - 2 - 6 = 2 (estoque mínimo).
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(5, nil).Once()
	reservaRepository.EXPECT().Atualizar(mock.Anything, reserva).Return(nil).Once()

	out, err := uc.Executar(context.Background(), AlterarQuantidadeReservaPecaInput{
		PecaID: 1, OrdemServicoID: 10, Quantidade: 6,
	})

	require.NoError(t, err)
	require.Equal(t, 6, out.Quantidade)
	require.Equal(t, 6, reserva.Quantidade())
	require.Equal(t, 1, runner.Calls)
}

func TestAlterarQuantidadeReservaPecaUseCase_RespeitaEstoqueMinimo(t *testing.T) {
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	uc := NewAlterarQuantidadeReservaPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	p := pecaComEstoque(t, 10, 2)
	reserva, err := domainreservapeca.NewReservaPeca(10, 1, 3)
	require.NoError(t, err)
	reserva.AtribuirID(5)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(mock.Anything, uint64(10), uint64(1)).Return(reserva, nil).Once()
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(5, nil).Once()

	_, err = uc.Executar(context.Background(), AlterarQuantidadeReservaPecaInput{
		PecaID: 1, OrdemServicoID: 10, Quantidade: 7,
	})

	require.ErrorIs(t, err, domain.ErrQuantidadeIndisponivelParaReserva)
	require.Equal(t, 3, reserva.Quantidade())
}
