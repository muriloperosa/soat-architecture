package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func atualizarPecaValida(t *testing.T) *domain.Peca {
	t.Helper()

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	p.AtribuirID(1)

	return p
}

func TestNewAtualizarPecaUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewAtualizarPecaUseCase(repository)

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
}

func TestAtualizarPecaUseCaseExecutarComSucesso(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarPecaUseCase(repository)

	p := atualizarPecaValida(t)

	input := AtualizarPecaInput{
		ID:            1,
		Nome:          "Peca 2",
		Marca:         "Marca 2",
		Descricao:     "Descricao 2",
		Preco:         200.0,
		EstoqueMinimo: 8,
	}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(p, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(
			ctx,
			mock.MatchedBy(func(p *domain.Peca) bool {
				return p != nil &&
					p.ID() == 1 &&
					p.Nome() == "Peca 2" &&
					p.Marca() == "Marca 2" &&
					p.Descricao() == "Descricao 2" &&
					p.Preco() == 200.0 &&
					p.EstoqueMinimo() == 8
			}),
		).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.NoError(t, err)

	require.Equal(t, uint64(1), output.ID)
	require.Equal(t, "Peca 2", output.Nome)
	require.Equal(t, "Marca 2", output.Marca)
	require.Equal(t, "Descricao 2", output.Descricao)
	require.Equal(t, 200.0, output.Preco)
	require.Equal(t, 8, output.EstoqueMinimo)
	require.Equal(t, 10, output.QuantidadeEmEstoque)
}

func TestAtualizarPecaUseCaseExecutarDeveRetornarErroAoBuscarPeca(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarPecaUseCase(repository)

	input := AtualizarPecaInput{ID: 999, Nome: "Peca 2", Marca: "Marca 2", Descricao: "Descricao 2", Preco: 200.0, EstoqueMinimo: 8}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(nil, domain.ErrPecaNaoEncontrada).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestAtualizarPecaUseCaseExecutarDeveRetornarErroDeValidacao(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarPecaUseCase(repository)

	p := atualizarPecaValida(t)

	input := AtualizarPecaInput{ID: 1, Nome: "", Marca: "Marca 2", Descricao: "Descricao 2", Preco: 200.0, EstoqueMinimo: 8}

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(p, nil).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrNomeObrigatorio)
}

func TestAtualizarPecaUseCaseExecutarDeveRetornarErroAoAtualizarRepository(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	useCase := NewAtualizarPecaUseCase(repository)

	p := atualizarPecaValida(t)

	input := AtualizarPecaInput{ID: 1, Nome: "Peca 2", Marca: "Marca 2", Descricao: "Descricao 2", Preco: 200.0, EstoqueMinimo: 8}

	erroRepository := errors.New("erro ao atualizar peca")

	repository.
		EXPECT().
		BuscarPorID(ctx, input.ID).
		Return(p, nil).
		Once()

	repository.
		EXPECT().
		Atualizar(ctx, p).
		Return(erroRepository).
		Once()

	output, err := useCase.Executar(ctx, input)

	require.Equal(t, PecaOutput{}, output)
	require.ErrorIs(t, err, erroRepository)
}
