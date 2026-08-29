package ordemservico_test

import (
	"context"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestInformarDiagnosticoUseCase_ComSucesso(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoEmDiagnostico(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	uc := app.NewInformarDiagnosticoUseCase(repository)
	output, err := uc.Executar(context.Background(), app.InformarDiagnosticoInput{
		OrdemServicoID: 42,
		Diagnostico:    "  Falha na bomba de combustível  ",
	})

	require.NoError(t, err)
	require.Equal(t, "Falha na bomba de combustível", output.Diagnostico)
	require.Equal(t, domain.StatusEmDiagnostico.String(), output.Status)
	require.Len(t, os.HistoricoStatus(), 2)
}

func TestInformarDiagnosticoUseCase_DiagnosticoVazio(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoEmDiagnostico(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	uc := app.NewInformarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.InformarDiagnosticoInput{OrdemServicoID: 42, Diagnostico: "   "})

	require.ErrorIs(t, err, domain.ErrDiagnosticoObrigatorio)
}

func TestInformarDiagnosticoUseCase_StatusInvalido(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebida(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	uc := app.NewInformarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.InformarDiagnosticoInput{
		OrdemServicoID: 42,
		Diagnostico:    "Falha na bomba",
	})

	require.ErrorIs(t, err, domain.ErrDiagnosticoStatusInvalido)
}

func ordemServicoEmDiagnostico(t *testing.T) *domain.OrdemServico {
	t.Helper()
	os := ordemServicoRecebida(t)
	require.NoError(t, os.IniciarDiagnostico(7))
	return os
}
