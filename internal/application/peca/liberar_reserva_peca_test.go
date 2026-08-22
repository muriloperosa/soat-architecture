package peca

import (
	"context"
	"errors"
	"testing"

	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/stretchr/testify/require"
)

func reservaExistente(t *testing.T, quantidade int) *domainreservapeca.ReservaPeca {
	t.Helper()

	r, err := domainreservapeca.NewReservaPeca(10, 1, quantidade)
	require.NoError(t, err)
	r.AtribuirID(5)

	return r
}

func TestNewLiberarReservaPecaUseCase(t *testing.T) {
	reservaRepository := reservamocks.NewRepository(t)

	useCase := NewLiberarReservaPecaUseCase(reservaRepository, transacaoFake{})

	require.NotNil(t, useCase)
	require.Equal(t, reservaRepository, useCase.reservaRepository)
}

func TestLiberarReservaPecaUseCaseExecutar_LiberacaoParcial(t *testing.T) {
	ctx := context.Background()
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewLiberarReservaPecaUseCase(reservaRepository, transacaoFake{})

	r := reservaExistente(t, 5)

	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(ctx, uint64(10), uint64(1)).Return(r, nil).Once()
	reservaRepository.EXPECT().Atualizar(ctx, r).Return(nil).Once()

	output, err := useCase.Executar(ctx, LiberarReservaPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 2})

	require.NoError(t, err)
	require.Equal(t, 3, output.Quantidade)
}

func TestLiberarReservaPecaUseCaseExecutar_LiberacaoTotal_RemoveAReserva(t *testing.T) {
	ctx := context.Background()
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewLiberarReservaPecaUseCase(reservaRepository, transacaoFake{})

	r := reservaExistente(t, 5)

	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(ctx, uint64(10), uint64(1)).Return(r, nil).Once()
	reservaRepository.EXPECT().Remover(ctx, uint64(10), uint64(1)).Return(nil).Once()

	output, err := useCase.Executar(ctx, LiberarReservaPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 5})

	require.NoError(t, err)
	require.Zero(t, output.Quantidade)
}

func TestLiberarReservaPecaUseCaseExecutar_ReservaNaoEncontrada_RetornaErro(t *testing.T) {
	ctx := context.Background()
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewLiberarReservaPecaUseCase(reservaRepository, transacaoFake{})

	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(ctx, uint64(10), uint64(1)).Return(nil, domainreservapeca.ErrReservaNaoEncontrada).Once()

	output, err := useCase.Executar(ctx, LiberarReservaPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 2})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, domainreservapeca.ErrReservaNaoEncontrada)
}

func TestLiberarReservaPecaUseCaseExecutar_QuantidadeMaiorQueReservada_RetornaErro(t *testing.T) {
	ctx := context.Background()
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewLiberarReservaPecaUseCase(reservaRepository, transacaoFake{})

	r := reservaExistente(t, 5)

	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(ctx, uint64(10), uint64(1)).Return(r, nil).Once()

	output, err := useCase.Executar(ctx, LiberarReservaPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 8})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, domainreservapeca.ErrQuantidadeSuperiorAReservada)
	require.Equal(t, 5, r.Quantidade())
}

func TestLiberarReservaPecaUseCaseExecutar_QuantidadeZeroOuNegativa_RetornaErro(t *testing.T) {
	ctx := context.Background()
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewLiberarReservaPecaUseCase(reservaRepository, transacaoFake{})

	output, err := useCase.Executar(ctx, LiberarReservaPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 0})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, domainreservapeca.ErrQuantidadeInvalida)
}

func TestLiberarReservaPecaUseCaseExecutar_ErroAoAtualizar_RetornaErro(t *testing.T) {
	ctx := context.Background()
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewLiberarReservaPecaUseCase(reservaRepository, transacaoFake{})

	r := reservaExistente(t, 5)
	erroBanco := errors.New("erro ao atualizar reserva")

	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(ctx, uint64(10), uint64(1)).Return(r, nil).Once()
	reservaRepository.EXPECT().Atualizar(ctx, r).Return(erroBanco).Once()

	output, err := useCase.Executar(ctx, LiberarReservaPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 2})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, erroBanco)
}
