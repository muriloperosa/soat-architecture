package servico_test

import (
	"context"
	"errors"
	"testing"

	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestCriarServicoUseCase_Executar_DadosValidos_CriaAtivo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewCriarServicoUseCase(repo)

	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(ctx context.Context, s *domainservico.Servico) { s.AtribuirID(1) }).
		Return(nil)

	out, err := uc.Executar(context.Background(), appservico.CriarServicoInput{
		Nome: "Troca de óleo", Descricao: "Troca de óleo e filtro", PrecoBase: 150.50, TempoEstimadoMinutos: 60, CriadoPor: 1,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(1), out.ID)
	require.True(t, out.Ativo)
	require.Equal(t, 150.50, out.PrecoBase)
	require.Equal(t, 60, out.TempoEstimadoMinutos)
	require.Equal(t, uint64(1), out.CriadoPor)
}

func TestCriarServicoUseCase_Executar_DadosInvalidos_PropagaErroDeValidacao(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewCriarServicoUseCase(repo)

	_, err := uc.Executar(context.Background(), appservico.CriarServicoInput{
		Nome: "", Descricao: "descrição", PrecoBase: 100, TempoEstimadoMinutos: 30, CriadoPor: 1,
	})

	require.ErrorIs(t, err, domainservico.ErrNomeObrigatorio)
}

func TestCriarServicoUseCase_Executar_ErroDoBancoAoSalvar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewCriarServicoUseCase(repo)

	repo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appservico.CriarServicoInput{
		Nome: "Troca de óleo", Descricao: "descrição", PrecoBase: 100, TempoEstimadoMinutos: 30, CriadoPor: 1,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
