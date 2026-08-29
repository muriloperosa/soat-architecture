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

func TestAtualizarServicoUseCase_Executar_ServicoExiste_AtualizaCampos(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(nil)

	out, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "Alinhamento", Descricao: "alinhamento e balanceamento", PrecoBase: 200.75, TempoEstimadoMinutos: 90,
	})

	require.NoError(t, err)
	require.Equal(t, "Alinhamento", out.Nome)
	require.Equal(t, 200.75, out.PrecoBase)
	require.Equal(t, 90, out.TempoEstimadoMinutos)
	require.Equal(t, uint64(1), out.CriadoPor)
	require.True(t, out.Ativo)
}

func TestAtualizarServicoUseCase_Executar_ServicoNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 99, Nome: "X", Descricao: "Y", PrecoBase: 10, TempoEstimadoMinutos: 15,
	})

	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestAtualizarServicoUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "Alinhamento", Descricao: "descrição", PrecoBase: 200, TempoEstimadoMinutos: 90,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtualizarServicoUseCase_Executar_DadosInvalidos_PropagaErroDeValidacao(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "", Descricao: "descrição", PrecoBase: 200, TempoEstimadoMinutos: 90,
	})

	require.ErrorIs(t, err, domainservico.ErrNomeObrigatorio)
}

func TestAtualizarServicoUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtualizarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID: 1, Nome: "Alinhamento", Descricao: "descrição", PrecoBase: 200, TempoEstimadoMinutos: 90,
	})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
