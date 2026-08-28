package relatorio_test

import (
	"context"
	"errors"
	"testing"
	"time"

	app "github.com/muriloperosa/soat-architecture/internal/application/relatorio"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
	relatoriomocks "github.com/muriloperosa/soat-architecture/internal/domain/relatorio/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func periodoValido() (time.Time, time.Time) {
	return time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC), time.Now()
}

func TestConsultarTransicaoStatusUseCase_ComSucesso(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	inicio, fim := periodoValido()

	repository.EXPECT().
		CalcularTransicaoStatus(mock.Anything, mock.MatchedBy(func(params domain.CalcularTransicaoStatusParams) bool {
			return params.FromStatus == domainordemservico.StatusRecebida &&
				params.ToStatus == domainordemservico.StatusEntregue
		})).
		Return(domain.TransicaoStatusResultado{
			TotalOrdens:   3,
			DuracaoMedia:  2 * time.Hour,
			DuracaoMinima: 1 * time.Hour,
			DuracaoMaxima: 3 * time.Hour,
		}, nil)

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	output, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: inicio,
		DataFim:    fim,
		DeStatus:   domainordemservico.StatusRecebida.String(),
		ParaStatus: domainordemservico.StatusEntregue.String(),
	})

	require.NoError(t, err)
	require.Equal(t, 3, output.TotalOrdensServico)
	require.Equal(t, 2*time.Hour, output.TempoMedio)
	require.Equal(t, 1*time.Hour, output.TempoMinimo)
	require.Equal(t, 3*time.Hour, output.TempoMaximo)
}

func TestConsultarTransicaoStatusUseCase_DeStatusInvalido(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	inicio, fim := periodoValido()

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	_, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: inicio,
		DataFim:    fim,
		DeStatus:   "INVALIDO",
		ParaStatus: domainordemservico.StatusEntregue.String(),
	})

	require.ErrorIs(t, err, domainordemservico.ErrStatusInvalido)
}

func TestConsultarTransicaoStatusUseCase_ParaStatusInvalido(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	inicio, fim := periodoValido()

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	_, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: inicio,
		DataFim:    fim,
		DeStatus:   domainordemservico.StatusRecebida.String(),
		ParaStatus: "INVALIDO",
	})

	require.ErrorIs(t, err, domainordemservico.ErrStatusInvalido)
}

func TestConsultarTransicaoStatusUseCase_StatusIguais(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	inicio, fim := periodoValido()

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	_, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: inicio,
		DataFim:    fim,
		DeStatus:   domainordemservico.StatusRecebida.String(),
		ParaStatus: domainordemservico.StatusRecebida.String(),
	})

	require.ErrorIs(t, err, domain.ErrTransicaoStatusIguais)
}

func TestConsultarTransicaoStatusUseCase_SemCaminhoValido(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	inicio, fim := periodoValido()

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	_, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: inicio,
		DataFim:    fim,
		DeStatus:   domainordemservico.StatusAprovada.String(),
		ParaStatus: domainordemservico.StatusRecebida.String(),
	})

	require.ErrorIs(t, err, domain.ErrTransicaoStatusSemCaminho)
}

func TestConsultarTransicaoStatusUseCase_PeriodoInvalido(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	_, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: time.Now(),
		DataFim:    time.Now().Add(-time.Hour),
		DeStatus:   domainordemservico.StatusRecebida.String(),
		ParaStatus: domainordemservico.StatusEntregue.String(),
	})

	require.ErrorIs(t, err, domain.ErrPeriodoInicioMaiorOuIgualFim)
}

func TestConsultarTransicaoStatusUseCase_ErroDoRepositorio(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	inicio, fim := periodoValido()
	erroBanco := errors.New("erro de conexão")

	repository.EXPECT().
		CalcularTransicaoStatus(mock.Anything, mock.Anything).
		Return(domain.TransicaoStatusResultado{}, erroBanco)

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	_, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: inicio,
		DataFim:    fim,
		DeStatus:   domainordemservico.StatusRecebida.String(),
		ParaStatus: domainordemservico.StatusEntregue.String(),
	})

	require.ErrorIs(t, err, erroBanco)
}

func TestConsultarTransicaoStatusUseCase_SemOrdensNoPeriodo(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	inicio, fim := periodoValido()

	repository.EXPECT().
		CalcularTransicaoStatus(mock.Anything, mock.Anything).
		Return(domain.TransicaoStatusResultado{}, nil)

	uc := app.NewConsultarTransicaoStatusUseCase(repository)
	output, err := uc.Executar(context.Background(), app.ConsultarTransicaoStatusInput{
		DataInicio: inicio,
		DataFim:    fim,
		DeStatus:   domainordemservico.StatusRecebida.String(),
		ParaStatus: domainordemservico.StatusEntregue.String(),
	})

	require.NoError(t, err)
	require.Zero(t, output.TotalOrdensServico)
	require.Zero(t, output.TempoMedio)
	require.Zero(t, output.TempoMinimo)
	require.Zero(t, output.TempoMaximo)
}
