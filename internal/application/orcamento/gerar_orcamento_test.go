package orcamento_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func ordemServicoExistente(t *testing.T, id uint64) *domainordemservico.OrdemServico {
	t.Helper()
	os, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 3)
	require.NoError(t, err)
	os.AtribuirID(id)
	return os
}

func TestGerarOrcamentoUseCase_ComSucesso(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(ordemServicoExistente(t, 42), nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)
	orcamentoRepo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(nil)

	uc := app.NewGerarOrcamentoUseCase(orcamentoRepo, osRepo)
	output, err := uc.Executar(context.Background(), app.GerarOrcamentoInput{
		OrdemServicoID: 42,
		Observacoes:    "orçamento inicial",
		UsuarioID:      7,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(42), output.OrdemServicoID)
	require.Equal(t, "orçamento inicial", output.Observacoes)
	require.Zero(t, output.ValorTotal)
}

func TestGerarOrcamentoUseCase_OSInexistente(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).
		Return(nil, domainordemservico.ErrOrdemServicoNaoEncontrada)

	uc := app.NewGerarOrcamentoUseCase(orcamentoRepo, osRepo)
	_, err := uc.Executar(context.Background(), app.GerarOrcamentoInput{OrdemServicoID: 999, UsuarioID: 7})

	require.ErrorIs(t, err, domainordemservico.ErrOrdemServicoNaoEncontrada)
}

func TestGerarOrcamentoUseCase_OrcamentoJaExiste(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(ordemServicoExistente(t, 42), nil)

	existente, err := domainorcamento.NewOrcamento(42, "", 7)
	require.NoError(t, err)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(existente, nil)

	uc := app.NewGerarOrcamentoUseCase(orcamentoRepo, osRepo)
	_, err = uc.Executar(context.Background(), app.GerarOrcamentoInput{OrdemServicoID: 42, UsuarioID: 7})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoJaExiste)
}

func TestGerarOrcamentoUseCase_ErroAoSalvar(t *testing.T) {
	orcamentoRepo := orcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	erroBanco := errors.New("banco indisponível")

	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(ordemServicoExistente(t, 42), nil)
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)
	orcamentoRepo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*orcamento.Orcamento")).Return(erroBanco)

	uc := app.NewGerarOrcamentoUseCase(orcamentoRepo, osRepo)
	_, err := uc.Executar(context.Background(), app.GerarOrcamentoInput{OrdemServicoID: 42, UsuarioID: 7})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}
