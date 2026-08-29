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

func TestListarServicosUseCase_Executar_RetornaCatalogo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	s1 := novoServico(t)
	s1.AtribuirID(1)
	s2, err := domainservico.NewServico("Alinhamento", "alinhamento e balanceamento", 200, 90, 1)
	require.NoError(t, err)
	s2.AtribuirID(2)

	repo.EXPECT().Listar(mock.Anything).Return([]*domainservico.Servico{s1, s2}, nil)

	out, err := uc.Executar(context.Background())
	require.NoError(t, err)
	require.Len(t, out, 2)
	require.Equal(t, "Troca de óleo", out[0].Nome)
	require.Equal(t, "Alinhamento", out[1].Nome)
}

func TestListarServicosUseCase_Executar_ListaVazia_RetornaSliceVazio(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	repo.EXPECT().Listar(mock.Anything).Return([]*domainservico.Servico{}, nil)

	out, err := uc.Executar(context.Background())
	require.NoError(t, err)
	require.Empty(t, out)
	require.NotNil(t, out)
}

func TestListarServicosUseCase_Executar_ErroDoBanco_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	repo.EXPECT().Listar(mock.Anything).Return(nil, errors.New("conexao recusada"))

	_, err := uc.Executar(context.Background())

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}
