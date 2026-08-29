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

func TestAtivarServicoUseCase_Executar_AtivaServicoInativo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	existente.Inativar()

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).
		Run(func(ctx context.Context, s *domainservico.Servico) { require.True(t, s.Ativo()) }).
		Return(nil)

	require.NoError(t, uc.Executar(context.Background(), 1))
}

func TestAtivarServicoUseCase_Executar_ServicoNaoExiste_RetornaNotFound(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(99)).Return(nil, domainservico.ErrServicoNaoEncontrado)

	err := uc.Executar(context.Background(), 99)
	require.ErrorIs(t, err, domainservico.ErrServicoNaoEncontrado)
}

func TestAtivarServicoUseCase_Executar_ErroDoBancoAoBuscar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestAtivarServicoUseCase_Executar_ErroDoBancoAoAtualizar_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewAtivarServicoUseCase(repo)

	existente := novoServico(t)
	existente.AtribuirID(1)
	existente.Inativar()

	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(existente, nil)
	repo.EXPECT().Atualizar(mock.Anything, mock.AnythingOfType("*servico.Servico")).Return(errors.New("conexao recusada"))

	err := uc.Executar(context.Background(), 1)

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
