package orcamento_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	pecamocks "github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func pecaExistente(t *testing.T) *domainpeca.Peca {
	t.Helper()
	p, err := domainpeca.NewPeca("Filtro de óleo", "Marca X", "Filtro de óleo do motor", 50.0, 10, 2, 3)
	require.NoError(t, err)
	p.AtribuirID(9)
	return p
}

func TestAdicionarPecaOrcamentoUseCase_ComSucesso(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(orcamentoVazio(t, 42), nil).Once()
	pecaRepo.EXPECT().BuscarPorID(mock.Anything, uint64(9)).Return(pecaExistente(t), nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).
		Run(func(_ context.Context, o *domainorcamento.Orcamento) {
			require.Len(t, o.ItensPeca(), 1)
			require.Equal(t, 150.0, o.ValorTotal())
		}).
		Return(nil)

	atualizado := orcamentoVazio(t, 42)
	require.NoError(t, atualizado.AdicionarItemPeca(9, "Filtro de óleo do motor", 3, 50.0))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(atualizado, nil).Once()

	uc := app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavel(t))
	output, err := uc.Executar(context.Background(), app.AdicionarPecaOrcamentoInput{
		OrdemServicoID: 42,
		PecaID:         9,
		Quantidade:     3,
	})

	require.NoError(t, err)
	require.Equal(t, 150.0, output.ValorTotal)
	require.Len(t, output.ItensPeca, 1)
	require.Equal(t, uint64(9), output.ItensPeca[0].PecaID)
	require.Equal(t, "Filtro de óleo do motor", output.ItensPeca[0].Descricao)
	require.Equal(t, 150.0, output.ItensPeca[0].Subtotal)
}

func TestAdicionarPecaOrcamentoUseCase_OrcamentoInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)

	uc := app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarPecaOrcamentoInput{OrdemServicoID: 42, PecaID: 9, Quantidade: 1})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoNaoEncontrado)
}

func TestAdicionarPecaOrcamentoUseCase_PecaInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	pecaRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainpeca.ErrPecaNaoEncontrada)

	uc := app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarPecaOrcamentoInput{OrdemServicoID: 42, PecaID: 999, Quantidade: 1})

	require.ErrorIs(t, err, domainpeca.ErrPecaNaoEncontrada)
}

func TestAdicionarPecaOrcamentoUseCase_QuantidadeInvalida(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	pecaRepo.EXPECT().BuscarPorID(mock.Anything, uint64(9)).Return(pecaExistente(t), nil)

	uc := app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarPecaOrcamentoInput{OrdemServicoID: 42, PecaID: 9, Quantidade: 0})

	require.ErrorIs(t, err, domainorcamento.ErrQuantidadeInvalida)
}

func TestAdicionarPecaOrcamentoUseCase_ErroAoAtualizar(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	erroBanco := errors.New("banco indisponível")

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	pecaRepo.EXPECT().BuscarPorID(mock.Anything, uint64(9)).Return(pecaExistente(t), nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(erroBanco)

	uc := app.NewAdicionarPecaOrcamentoUseCase(orcamentoRepo, pecaRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarPecaOrcamentoInput{OrdemServicoID: 42, PecaID: 9, Quantidade: 1})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}
