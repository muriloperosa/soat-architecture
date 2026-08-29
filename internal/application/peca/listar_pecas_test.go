package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func novaPecaParaListagem(t *testing.T) *domain.Peca {
	t.Helper()

	peca, err := domain.NewPeca("Pastilha de Freio", "Bosch", "Pastilha dianteira", 149.90, 25, 5, 1)
	require.NoError(t, err)

	peca.AtribuirID(1)

	return peca
}

func TestNewListarPecasUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)

	useCase := NewListarPecasUseCase(repository)

	require.NotNil(t, useCase)
}

func TestListarPecasUseCase_Executar_ComSucesso(t *testing.T) {
	repository := mocks.NewRepository(t)
	useCase := NewListarPecasUseCase(repository)

	ctx := context.Background()

	params := query.Params{
		Offset:    10,
		Limit:     20,
		Order:     "nome",
		Direction: query.DirectionDESC,
	}

	peca := novaPecaParaListagem(t)

	repository.
		EXPECT().
		Listar(ctx, params).
		Return(
			query.Page[*domain.Peca]{
				Items: []*domain.Peca{
					peca,
				},
				Total:     42,
				Offset:    10,
				Limit:     20,
				Order:     "nome",
				Direction: query.DirectionDESC,
			},
			nil,
		).
		Once()

	page, err := useCase.Executar(ctx, params)

	require.NoError(t, err)

	require.Equal(t, int64(42), page.Total)
	require.Equal(t, 10, page.Offset)
	require.Equal(t, 20, page.Limit)
	require.Equal(t, "nome", page.Order)
	require.Equal(t, query.DirectionDESC, page.Direction)

	require.Len(t, page.Items, 1)

	item := page.Items[0]

	require.Equal(t, peca.ID(), item.ID)
	require.Equal(t, peca.Codigo(), item.Codigo)
	require.Equal(t, peca.Nome(), item.Nome)
	require.Equal(t, peca.Marca(), item.Marca)
	require.Equal(t, peca.Descricao(), item.Descricao)
	require.Equal(t, peca.Preco(), item.Preco)
	require.Equal(t, peca.QuantidadeEmEstoque(), item.QuantidadeEmEstoque)
	require.Equal(t, peca.EstoqueMinimo(), item.EstoqueMinimo)
	require.Equal(t, peca.CriadoPor(), item.CriadoPor)
	require.Equal(t, peca.Ativo(), item.Ativo)
}

func TestListarPecasUseCase_Executar_ListaVazia(t *testing.T) {
	repository := mocks.NewRepository(t)
	useCase := NewListarPecasUseCase(repository)

	ctx := context.Background()

	params := query.Params{}

	repository.
		EXPECT().
		Listar(ctx, params).
		Return(
			query.Page[*domain.Peca]{
				Items:     []*domain.Peca{},
				Total:     0,
				Offset:    0,
				Limit:     20,
				Order:     "id",
				Direction: query.DirectionASC,
			},
			nil,
		).
		Once()

	page, err := useCase.Executar(ctx, params)

	require.NoError(t, err)

	require.NotNil(t, page.Items)
	require.Empty(t, page.Items)

	require.Equal(t, int64(0), page.Total)
	require.Equal(t, 0, page.Offset)
	require.Equal(t, 20, page.Limit)
	require.Equal(t, "id", page.Order)
	require.Equal(t, query.DirectionASC, page.Direction)
}

func TestListarPecasUseCase_Executar_PreservaAppError(t *testing.T) {
	repository := mocks.NewRepository(t)
	useCase := NewListarPecasUseCase(repository)

	ctx := context.Background()

	params := query.Params{Order: "campo_invalido"}

	appErr := shared.NewValidationError("campo de ordenação inválido")

	repository.
		EXPECT().
		Listar(ctx, params).
		Return(query.Page[*domain.Peca]{}, appErr).
		Once()

	page, err := useCase.Executar(ctx, params)

	require.Error(t, err)
	require.ErrorIs(t, err, appErr)

	require.Empty(t, page.Items)
}

func TestListarPecasUseCase_Executar_ConverteErroDoRepositoryParaInternalError(t *testing.T) {
	repository := mocks.NewRepository(t)
	useCase := NewListarPecasUseCase(repository)

	ctx := context.Background()

	params := query.Params{}

	erroBanco := errors.New("erro de banco")

	repository.
		EXPECT().
		Listar(ctx, params).
		Return(query.Page[*domain.Peca]{}, erroBanco).
		Once()

	page, err := useCase.Executar(ctx, params)

	require.Error(t, err)
	require.Empty(t, page.Items)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)

	require.ErrorIs(t, err, erroBanco)
}
