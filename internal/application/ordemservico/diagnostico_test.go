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

func TestIniciarDiagnosticoUseCase_ComSucesso(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebida(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).
		Run(func(_ context.Context, atualizada *domain.OrdemServico) {
			require.Equal(t, domain.StatusEmDiagnostico, atualizada.Status())
			require.Len(t, atualizada.HistoricoStatus(), 2)
			require.Equal(t, uint64(7), atualizada.HistoricoStatus()[1].AlteradoPor())
			require.Equal(t, "Diagnóstico iniciado", atualizada.HistoricoStatus()[1].Motivo())
		}).
		Return(nil)

	uc := app.NewIniciarDiagnosticoUseCase(repository)
	output, err := uc.Executar(context.Background(), app.IniciarDiagnosticoInput{
		OrdemServicoID: 42,
		UsuarioID:      7,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(42), output.ID)
	require.Equal(t, domain.StatusEmDiagnostico.String(), output.Status)
}

func TestIniciarDiagnosticoUseCase_OSInexistente(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	uc := app.NewIniciarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.IniciarDiagnosticoInput{OrdemServicoID: 999, UsuarioID: 7})

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}

func TestIniciarDiagnosticoUseCase_TransicaoInvalida(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebida(t)
	require.NoError(t, os.IniciarDiagnostico(7, "Diagnóstico iniciado"))
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	uc := app.NewIniciarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.IniciarDiagnosticoInput{OrdemServicoID: 42, UsuarioID: 7})

	require.ErrorIs(t, err, domain.ErrTransicaoStatusInvalida)
}

func TestIniciarDiagnosticoUseCase_ErroAoPersistir(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebida(t)
	erroBanco := errors.New("banco indisponível")
	repository.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(erroBanco)

	uc := app.NewIniciarDiagnosticoUseCase(repository)
	_, err := uc.Executar(context.Background(), app.IniciarDiagnosticoInput{OrdemServicoID: 42, UsuarioID: 7})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}

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

func ordemServicoRecebida(t *testing.T) *domain.OrdemServico {
	t.Helper()
	os, err := domain.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 3)
	require.NoError(t, err)
	os.AtribuirID(42)
	return os
}

func ordemServicoEmDiagnostico(t *testing.T) *domain.OrdemServico {
	t.Helper()
	os := ordemServicoRecebida(t)
	require.NoError(t, os.IniciarDiagnostico(7, "Diagnóstico iniciado"))
	return os
}
