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

func TestBuscarServicoUseCase_Executar_ServicoExiste_RetornaDados(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewBuscarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)

	out, err := uc.Executar(context.Background(), 1)
	require.NoError(t, err)
	require.Equal(t, uint64(1), out.ID)
	require.Equal(t, "Troca de óleo", out.Nome)
}

func TestBuscarServicoUseCase_Executar_ServicoNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewBuscarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	_, err := uc.Executar(context.Background(), 99)
	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestBuscarServicoUseCase_Executar_ErroDoBanco_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewBuscarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
