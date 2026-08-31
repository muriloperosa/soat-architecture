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

func TestNewConsultarPecaPorIDUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)

	useCase := NewConsultarPecaPorIDUseCase(repository, reservaRepository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
	require.Equal(t, reservaRepository, useCase.reservaRepository)
}

func TestConsultarPecaPorIDUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewConsultarPecaPorIDUseCase(repository, reservaRepository)

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)
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

	resultado, err := useCase.Executar(ctx, 1)

	require.NoError(t, err)
	require.Equal(t, uint64(1), resultado.ID)
	require.Equal(t, "Peca 1", resultado.Nome)
	require.Equal(t, 10, resultado.QuantidadeEmEstoque)
	require.Equal(t, 4, resultado.QuantidadeReservada)
	require.Equal(t, 6, resultado.QuantidadeDisponivel)
}

func TestConsultarPecaPorIDUseCaseExecutarDeveRetornarErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewConsultarPecaPorIDUseCase(repository, reservaRepository)

	repository.
		EXPECT().
		BuscarPorID(ctx, uint64(999)).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	resultado, err := useCase.Executar(ctx, 999)

	require.Equal(t, PecaOutput{}, resultado)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestConsultarPecaPorIDUseCaseExecutarDeveRetornarErroAoSomarReservado(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewConsultarPecaPorIDUseCase(repository, reservaRepository)

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)
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

	resultado, err := useCase.Executar(ctx, 1)

	require.Equal(t, PecaOutput{}, resultado)
	require.ErrorIs(t, err, erroRepository)
}
