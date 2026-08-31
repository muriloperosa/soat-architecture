package orcamento_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	servicomocks "github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func orcamentoVazio(t *testing.T, ordemServicoID uint64) *domainorcamento.Orcamento {
	t.Helper()
	o, err := domainorcamento.NewOrcamento(ordemServicoID, "", 3)
	require.NoError(t, err)
	o.AtribuirID(1)
	return o
}

func servicoAtivo(t *testing.T) *domainservico.Servico {
	t.Helper()
	s, err := domainservico.NewServico("Troca de óleo", "Troca de óleo do motor", 100.0, 60, 3)
	require.NoError(t, err)
	s.AtribuirID(5)
	return s
}

func TestAdicionarServicoOrcamentoUseCase_ComSucesso(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(orcamentoVazio(t, 42), nil).Once()
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(5)).Return(servicoAtivo(t), nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).
		Run(func(_ context.Context, o *domainorcamento.Orcamento) {
			require.Len(t, o.ItensServico(), 1)
			require.Equal(t, 200.0, o.ValorTotal())
		}).
		Return(nil)

	atualizado := orcamentoVazio(t, 42)
	require.NoError(t, atualizado.AdicionarItemServico(5, 2, 100.0, 60))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(atualizado, nil).Once()

	uc := app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavel(t))
	output, err := uc.Executar(context.Background(), app.AdicionarServicoOrcamentoInput{
		OrdemServicoID: 42,
		ServicoID:      5,
		Quantidade:     2,
	})

	require.NoError(t, err)
	require.Equal(t, 200.0, output.ValorTotal)
	require.Len(t, output.ItensServico, 1)
	require.Equal(t, uint64(5), output.ItensServico[0].ServicoID)
	require.Equal(t, 200.0, output.ItensServico[0].Subtotal)
}

func TestAdicionarServicoOrcamentoUseCase_OrcamentoInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)

	uc := app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarServicoOrcamentoInput{OrdemServicoID: 42, ServicoID: 5, Quantidade: 1})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoNaoEncontrado)
}

func TestAdicionarServicoOrcamentoUseCase_ServicoInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	uc := app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarServicoOrcamentoInput{OrdemServicoID: 42, ServicoID: 999, Quantidade: 1})

	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestAdicionarServicoOrcamentoUseCase_ServicoInativo(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	servicoInativo := servicoAtivo(t)
	servicoInativo.Inativar()

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(5)).Return(servicoInativo, nil)

	uc := app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarServicoOrcamentoInput{OrdemServicoID: 42, ServicoID: 5, Quantidade: 1})

	require.ErrorIs(t, err, domainorcamento.ErrServicoInativo)
}

func TestAdicionarServicoOrcamentoUseCase_QuantidadeInvalida(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(5)).Return(servicoAtivo(t), nil)

	uc := app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarServicoOrcamentoInput{OrdemServicoID: 42, ServicoID: 5, Quantidade: 0})

	require.ErrorIs(t, err, domainorcamento.ErrQuantidadeInvalida)
}

func TestAdicionarServicoOrcamentoUseCase_ErroAoAtualizar(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	servicoRepo := servicomocks.NewServicoRepository(t)
	erroBanco := errors.New("banco indisponível")

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	servicoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(5)).Return(servicoAtivo(t), nil)
	orcamentoRepo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(erroBanco)

	uc := app.NewAdicionarServicoOrcamentoUseCase(orcamentoRepo, servicoRepo, ordemServicoRepoEditavel(t))
	_, err := uc.Executar(context.Background(), app.AdicionarServicoOrcamentoInput{OrdemServicoID: 42, ServicoID: 5, Quantidade: 1})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}
