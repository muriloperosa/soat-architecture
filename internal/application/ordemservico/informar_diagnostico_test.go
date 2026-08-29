package ordemservico_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
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

func TestInformarDiagnosticoUseCase_OSInexistente(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	uc := app.NewInformarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.InformarDiagnosticoInput{OrdemServicoID: 999, Diagnostico: "Falha"})

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}

func TestInformarDiagnosticoUseCase_ErroAoBuscar(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	erroBanco := errors.New("banco indisponível")
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(nil, erroBanco)

	uc := app.NewInformarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.InformarDiagnosticoInput{OrdemServicoID: 42, Diagnostico: "Falha"})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}

func TestInformarDiagnosticoUseCase_OSNula(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(nil, nil)

	uc := app.NewInformarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.InformarDiagnosticoInput{OrdemServicoID: 42, Diagnostico: "Falha"})

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}

func TestInformarDiagnosticoUseCase_ErroAoPersistir(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoEmDiagnostico(t)
	erroBanco := errors.New("banco indisponível")
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(erroBanco)

	uc := app.NewInformarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.InformarDiagnosticoInput{
		OrdemServicoID: 42,
		Diagnostico:    "Falha na bomba",
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
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
