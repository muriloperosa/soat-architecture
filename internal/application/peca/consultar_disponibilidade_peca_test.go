package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/stretchr/testify/require"
)

func TestNewConsultarDisponibilidadeUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)

	useCase := NewConsultarDisponibilidadeUseCase(repository, reservaRepository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
	require.Equal(t, reservaRepository, useCase.reservaRepository)
}

func TestConsultarDisponibilidadeUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewConsultarDisponibilidadeUseCase(repository, reservaRepository)

	p := novaPecaValida(t)
	p.AtribuirID(1)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(p, nil).
		Once()

	reservaRepository.
		EXPECT().
		SomarQuantidadeReservada(ctx, uint64(1)).
		Return(4, nil).
		Once()

	output, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, uint64(1), output.PecaID)
	require.Equal(t, 10, output.QuantidadeEmEstoque)
	require.Equal(t, 4, output.QuantidadeReservada)
	require.Equal(t, 6, output.QuantidadeDisponivel)
}

func TestConsultarDisponibilidadeUseCaseExecutarDeveRetornarErroAoBuscarPeca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewConsultarDisponibilidadeUseCase(repository, reservaRepository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	output, err := useCase.Executar(ctx, 999)

	require.Equal(t, DisponibilidadePecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestConsultarDisponibilidadeUseCaseExecutarDeveRetornarErroAoSomarReservado(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewConsultarDisponibilidadeUseCase(repository, reservaRepository)

	p := novaPecaValida(t)
	p.AtribuirID(1)

	erroRepository := errors.New("erro ao somar reservas")

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(1)).
		Return(p, nil).
		Once()

	reservaRepository.
		EXPECT().
		SomarQuantidadeReservada(ctx, uint64(1)).
		Return(0, erroRepository).
		Once()

	output, err := useCase.Executar(ctx, 1)

	require.Equal(t, DisponibilidadePecaOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
