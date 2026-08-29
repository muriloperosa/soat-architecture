package ordemservico_test

import (
	"context"
	"errors"
	"testing"
	"time"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestEntregarOrdemServicoUseCase_ComSucesso(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoFinalizada(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).
		Run(func(_ context.Context, atualizada *domain.OrdemServico) {
			require.Equal(t, domain.StatusEntregue, atualizada.Status())
			require.Len(t, atualizada.HistoricoStatus(), 3)
			require.Equal(t, uint64(7), atualizada.HistoricoStatus()[2].AlteradoPor())
			require.Empty(t, atualizada.HistoricoStatus()[2].Motivo())
		}).
		Return(nil)

	uc := app.NewEntregarOrdemServicoUseCase(repository)
	output, err := uc.Executar(context.Background(), app.EntregarOrdemServicoInput{
		OrdemServicoID: 42,
		UsuarioID:      7,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(42), output.ID)
	require.Equal(t, domain.StatusEntregue.String(), output.Status)
}

func TestEntregarOrdemServicoUseCase_OSInexistente(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	uc := app.NewEntregarOrdemServicoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.EntregarOrdemServicoInput{OrdemServicoID: 999, UsuarioID: 7})

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}

func TestEntregarOrdemServicoUseCase_TransicaoInvalida(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebida(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	uc := app.NewEntregarOrdemServicoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.EntregarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	require.ErrorIs(t, err, domain.ErrTransicaoStatusInvalida)
}

func TestEntregarOrdemServicoUseCase_ErroAoPersistir(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoFinalizada(t)
	erroBanco := errors.New("banco indisponível")
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(erroBanco)

	uc := app.NewEntregarOrdemServicoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.EntregarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}

func ordemServicoFinalizada(t *testing.T) *domain.OrdemServico {
	t.Helper()
	numero, err := domain.NewNumeroOrdemServico("OS-20260827-a1b2c3d4e5f6")
	require.NoError(t, err)

	cadastro := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	atualizacao := cadastro.Add(3 * time.Hour)
	historico := []domain.HistoricoStatus{
		domain.ReidratarHistoricoStatus(1, 42, domain.StatusRecebida, cadastro, 3, ""),
		domain.ReidratarHistoricoStatus(2, 42, domain.StatusFinalizada, atualizacao, 3, ""),
	}

	return domain.ReidratarOrdemServico(
		42,
		numero,
		10,
		20,
		52_300,
		domain.StatusFinalizada,
		"",
		"",
		3,
		historico,
		cadastro,
		atualizacao,
	)
}
