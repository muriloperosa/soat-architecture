package orcamento_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func orcamentoComItemPeca(t *testing.T, ordemServicoID uint64) *domainorcamento.Orcamento {
	t.Helper()
	o := orcamentoVazio(t, ordemServicoID)
	require.NoError(t, o.AdicionarItemPeca(9, "Filtro de óleo", 3, 50.0))
	item := o.ItensPeca()[0]

	itens := []domainorcamento.ItemPeca{
		domainorcamento.ReidratarItemPeca(1, o.ID(), item.PecaID(), item.Descricao(), item.Quantidade(), item.Valor()),
	}
	return domainorcamento.ReidratarOrcamento(
		o.ID(), ordemServicoID,
		nil, itens,
		o.ValorItemServicos(), o.ValorItemPecas(), o.ValorTotal(),
		o.Observacoes(), o.CriadoPor(), o.CriadoEm(), o.AtualizadoEm(),
	)
}

func TestRemoverPecaOrcamentoUseCase_ComSucesso(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(orcamentoComItemPeca(t, 42), nil).Once()
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).
		Run(func(_ context.Context, o *domainorcamento.Orcamento) {
			require.Empty(t, o.ItensPeca())
			require.Zero(t, o.ValorTotal())
		}).
		Return(nil)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(orcamentoVazio(t, 42), nil).Once()

	uc := app.NewRemoverPecaOrcamentoUseCase(orcamentoRepo)
	output, err := uc.Executar(context.Background(), app.RemoverPecaOrcamentoInput{OrdemServicoID: 42, ItemPecaID: 1})

	require.NoError(t, err)
	require.Empty(t, output.ItensPeca)
}

func TestRemoverPecaOrcamentoUseCase_OrcamentoInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)

	uc := app.NewRemoverPecaOrcamentoUseCase(orcamentoRepo)
	_, err := uc.Executar(context.Background(), app.RemoverPecaOrcamentoInput{OrdemServicoID: 42, ItemPecaID: 1})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoNaoEncontrado)
}

func TestRemoverPecaOrcamentoUseCase_ItemInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)

	uc := app.NewRemoverPecaOrcamentoUseCase(orcamentoRepo)
	_, err := uc.Executar(context.Background(), app.RemoverPecaOrcamentoInput{OrdemServicoID: 42, ItemPecaID: 999})

	require.ErrorIs(t, err, domainorcamento.ErrItemPecaNaoEncontrado)
}

func TestRemoverPecaOrcamentoUseCase_ErroAoAtualizar(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	erroBanco := errors.New("banco indisponível")

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoComItemPeca(t, 42), nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(erroBanco)

	uc := app.NewRemoverPecaOrcamentoUseCase(orcamentoRepo)
	_, err := uc.Executar(context.Background(), app.RemoverPecaOrcamentoInput{OrdemServicoID: 42, ItemPecaID: 1})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}
