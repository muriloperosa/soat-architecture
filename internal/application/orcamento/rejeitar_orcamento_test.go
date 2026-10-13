package orcamento_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRejeitarOrcamentoUseCase_RegistraMotivoNoHistorico(t *testing.T) {
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	o := osAguardandoAprovacao(t, 10)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	repo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewRejeitarOrcamentoUseCase(repo)
	out, err := uc.Executar(context.Background(), app.RejeitarOrcamentoInput{
		OrdemServicoID: 42,
		ClienteID:      10,
		Motivo:         "Valor acima do esperado",
	})

	require.NoError(t, err)
	require.Equal(t, "REJEITADA", out.Status)
	historico := o.HistoricoStatus()
	require.Equal(t, "Valor acima do esperado", historico[len(historico)-1].Motivo())
	require.Equal(t, uint64(2), historico[len(historico)-1].AlteradoPor())
}

func TestRejeitarOrcamentoUseCase_ErrosDeBuscaEAtualizacao(t *testing.T) {
	t.Run("ordem de serviço não encontrada", func(t *testing.T) {
		repo := ordemservicomocks.NewOrdemServicoRepository(t)
		repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(nil, domainordemservico.ErrOrdemServicoNaoEncontrada)

		uc := app.NewRejeitarOrcamentoUseCase(repo)
		_, err := uc.Executar(context.Background(), app.RejeitarOrcamentoInput{OrdemServicoID: 42, ClienteID: 10})
		require.ErrorIs(t, err, domainordemservico.ErrOrdemServicoNaoEncontrada)
	})

	t.Run("falha de infraestrutura ao buscar", func(t *testing.T) {
		repo := ordemservicomocks.NewOrdemServicoRepository(t)
		cause := errors.New("banco indisponível")
		repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(nil, cause)

		uc := app.NewRejeitarOrcamentoUseCase(repo)
		_, err := uc.Executar(context.Background(), app.RejeitarOrcamentoInput{OrdemServicoID: 42, ClienteID: 10})
		var appErr *shared.AppError
		require.ErrorAs(t, err, &appErr)
		require.ErrorIs(t, err, cause)
	})

	t.Run("falha ao persistir", func(t *testing.T) {
		repo := ordemservicomocks.NewOrdemServicoRepository(t)
		os := osAguardandoAprovacao(t, 10)
		cause := errors.New("banco indisponível")
		repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
		repo.EXPECT().Atualizar(mock.Anything, os).Return(cause)

		uc := app.NewRejeitarOrcamentoUseCase(repo)
		_, err := uc.Executar(context.Background(), app.RejeitarOrcamentoInput{OrdemServicoID: 42, ClienteID: 10, Motivo: "Não aprovado"})
		var appErr *shared.AppError
		require.ErrorAs(t, err, &appErr)
		require.ErrorIs(t, err, cause)
	})
}
