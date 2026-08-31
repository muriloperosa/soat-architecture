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

func orcamentoComItemServico(t *testing.T, ordemServicoID uint64) *domainorcamento.Orcamento {
	t.Helper()
	o := orcamentoVazio(t, ordemServicoID)
	require.NoError(t, o.AdicionarItemServico(5, 2, 100.0, 60))
	item := o.ItensServico()[0]

	itens := []domainorcamento.ItemServico{
		domainorcamento.ReidratarItemServico(1, o.ID(), item.ServicoID(), item.Quantidade(), item.Valor(), item.TempoEstimado()),
	}
	return domainorcamento.ReidratarOrcamento(
		o.ID(), ordemServicoID,
		itens, nil,
		o.ValorItemServicos(), o.ValorItemPecas(), o.ValorTotal(),
		o.Observacoes(), o.CriadoPor(), o.CriadoEm(), o.AtualizadoEm(),
	)
}

func TestRemoverServicoOrcamentoUseCase_ComSucesso(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(orcamentoComItemServico(t, 42), nil).Once()
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).
		Run(func(_ context.Context, o *domainorcamento.Orcamento) {
			require.Empty(t, o.ItensServico())
			require.Zero(t, o.ValorTotal())
		}).
		Return(nil)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(orcamentoVazio(t, 42), nil).Once()

	uc := app.NewRemoverServicoOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavel(t))
	output, err := uc.Executar(context.Background(), app.RemoverServicoOrcamentoInput{OrdemServicoID: 42, ItemServicoID: 1})

	require.NoError(t, err)
	require.Empty(t, output.ItensServico)
}

func TestRemoverServicoOrcamentoUseCase_OrcamentoInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)

	uc := app.NewRemoverServicoOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.RemoverServicoOrcamentoInput{OrdemServicoID: 42, ItemServicoID: 1})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoNaoEncontrado)
}

func TestRemoverServicoOrcamentoUseCase_ItemInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)

	uc := app.NewRemoverServicoOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.RemoverServicoOrcamentoInput{OrdemServicoID: 42, ItemServicoID: 999})

	require.ErrorIs(t, err, domainorcamento.ErrItemServicoNaoEncontrado)
}

func TestRemoverServicoOrcamentoUseCase_ErroAoAtualizar(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	erroBanco := errors.New("banco indisponível")

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoComItemServico(t, 42), nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(erroBanco)

	uc := app.NewRemoverServicoOrcamentoUseCase(orcamentoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.RemoverServicoOrcamentoInput{OrdemServicoID: 42, ItemServicoID: 1})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}
