package servico_test

import (
	"context"
	"errors"
	"testing"

	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/servico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestListarServicosUseCase_Executar_RetornaCatalogo(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	ctx := context.Background()

	params := query.Params{
		Offset:    0,
		Limit:     20,
		Order:     "nome",
		Direction: query.DirectionASC,
	}

	s1 := novoServico(t)
	s1.AtribuirID(1)

	s2, err := domainservico.NewServico(
		"Alinhamento",
		"alinhamento e balanceamento",
		200,
		90,
		1,
	)
	require.NoError(t, err)

	s2.AtribuirID(2)

	repo.
		EXPECT().
		Listar(ctx, params).
		Return(
			query.Page[*domainservico.Servico]{
				Items: []*domainservico.Servico{
					s1,
					s2,
				},
				Total:     2,
				Offset:    0,
				Limit:     20,
				Order:     "nome",
				Direction: query.DirectionASC,
			},
			nil,
		).
		Once()

	out, err := uc.Executar(ctx, params)

	require.NoError(t, err)

	require.Equal(t, int64(2), out.Total)
	require.Equal(t, 0, out.Offset)
	require.Equal(t, 20, out.Limit)
	require.Equal(t, "nome", out.Order)
	require.Equal(t, query.DirectionASC, out.Direction)

	require.Len(t, out.Items, 2)

	require.Equal(t, "Troca de óleo", out.Items[0].Nome)
	require.Equal(t, "Alinhamento", out.Items[1].Nome)
}

func TestListarServicosUseCase_Executar_ListaVazia_RetornaPaginaVazia(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	ctx := context.Background()

	params := query.Params{}

	repo.
		EXPECT().
		Listar(ctx, params).
		Return(
			query.Page[*domainservico.Servico]{
				Items:     []*domainservico.Servico{},
				Total:     0,
				Offset:    0,
				Limit:     20,
				Order:     "id",
				Direction: query.DirectionASC,
			},
			nil,
		).
		Once()

	out, err := uc.Executar(ctx, params)

	require.NoError(t, err)

	require.NotNil(t, out.Items)
	require.Empty(t, out.Items)

	require.Equal(t, int64(0), out.Total)
	require.Equal(t, 0, out.Offset)
	require.Equal(t, 20, out.Limit)
	require.Equal(t, "id", out.Order)
	require.Equal(t, query.DirectionASC, out.Direction)
}

func TestListarServicosUseCase_Executar_ErroDoBanco_RetornaInternalError(t *testing.T) {
	repo := mocks.NewServicoRepository(t)
	uc := appservico.NewListarServicosUseCase(repo)

	ctx := context.Background()

	params := query.Params{}

	erroBanco := errors.New("conexao recusada")

	repo.
		EXPECT().
		Listar(ctx, params).
		Return(
			query.Page[*domainservico.Servico]{},
			erroBanco,
		).
		Once()

	out, err := uc.Executar(ctx, params)

	require.Error(t, err)
	require.Empty(t, out.Items)

	var appErr *shared.AppError

	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}
