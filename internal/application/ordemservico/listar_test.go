package ordemservico_test

import (
	"context"
	"errors"
	"testing"
	"time"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListarOrdensServicoUseCase_Executar_ComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	ctx := context.Background()
	os := ordemServicoConsultaValida(t)

	input := app.ListarOrdensServicoInput{ParamsInput: appquery.ParamsInput{
		Page: 2, Order: "data_cadastro", Direction: "DESC",
		Filters: []appquery.FilterInput{{Field: "status", Operator: "auto", Value: "RECEBIDA"}},
	}}
	expectedParams := query.Params{
		Page: 2, Order: "data_cadastro", Direction: query.DirectionDESC,
		Filters: []query.Filter{{Field: "status", Operator: query.OperatorAuto, Value: "RECEBIDA"}},
	}

	repository.EXPECT().Listar(ctx, expectedParams).Return(query.Page[*domain.OrdemServico]{
		Items: []*domain.OrdemServico{os}, Total: 25, Page: 2, PageSize: 20, TotalPages: 2,
		Order: "data_cadastro", Direction: query.DirectionDESC,
	}, nil).Once()

	out, err := app.NewListarOrdensServicoUseCase(repository).Executar(ctx, input)
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Equal(t, int64(25), out.Total)
	require.Equal(t, 2, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Equal(t, 2, out.TotalPages)
	require.Equal(t, "data_cadastro", out.Order)
	require.Equal(t, "DESC", out.Direction)
	require.Equal(t, os.Numero().String(), out.Items[0].Numero)
}

func TestListarOrdensServicoUseCase_Executar_ListaVazia(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	repository.EXPECT().Listar(mock.Anything, query.Params{Page: 1}).Return(query.Page[*domain.OrdemServico]{
		Items: []*domain.OrdemServico{}, Total: 0, Page: 1, PageSize: 20, TotalPages: 0,
		Order: "id", Direction: query.DirectionASC,
	}, nil)

	out, err := app.NewListarOrdensServicoUseCase(repository).Executar(context.Background(), app.ListarOrdensServicoInput{
		ParamsInput: appquery.ParamsInput{Page: 1},
	})
	require.NoError(t, err)
	require.NotNil(t, out.Items)
	require.Empty(t, out.Items)
	require.Equal(t, 1, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Zero(t, out.TotalPages)
}

func TestListarOrdensServicoUseCase_Executar_PreservaAppError(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	appErr := shared.NewValidationError("filtro inválido")
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(query.Page[*domain.OrdemServico]{}, appErr)

	_, err := app.NewListarOrdensServicoUseCase(repository).Executar(context.Background(), app.ListarOrdensServicoInput{})
	require.ErrorIs(t, err, appErr)
}

func TestListarOrdensServicoUseCase_Executar_EncapsulaErroInfra(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(query.Page[*domain.OrdemServico]{}, errors.New("banco indisponível"))

	_, err := app.NewListarOrdensServicoUseCase(repository).Executar(context.Background(), app.ListarOrdensServicoInput{})
	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func ordemServicoConsultaValida(t *testing.T) *domain.OrdemServico {
	t.Helper()
	numero, err := domain.NewNumeroOrdemServico("OS-20260830-a1b2c3d4e5f6")
	require.NoError(t, err)
	agora := time.Date(2026, time.August, 30, 12, 0, 0, 0, time.UTC)
	historico := domain.ReidratarHistoricoStatus(1, 42, domain.StatusRecebida, agora, 7, "")
	return domain.ReidratarOrdemServico(42, numero, 10, 20, 52300, domain.StatusRecebida, "", "Ruído", 7, []domain.HistoricoStatus{historico}, agora, agora)
}
