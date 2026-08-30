package ordemservico_test

import (
	"context"
	"errors"
	"testing"
	"time"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	orcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestListarOrdensServicoUseCase_Executar_ComSucesso(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	orcamentoRepository := orcamentomocks.NewOrcamentoRepository(t)
	ctx := context.Background()
	os := ordemServicoConsultaValida(t)
	orcamento := orcamentoConsultaValido()

	input := app.ListarOrdensServicoInput{
		ParamsInput: appquery.ParamsInput{
			Page: 2, Order: "data_cadastro", Direction: "DESC",
			Filters: []appquery.FilterInput{{Field: "status", Operator: "auto", Value: "RECEBIDA"}},
		},
		SolicitanteID:   7,
		TipoSolicitante: domainauth.TipoInterno,
	}
	expectedParams := query.Params{
		Page: 2, Order: "data_cadastro", Direction: query.DirectionDESC,
		Filters: []query.Filter{{Field: "status", Operator: query.OperatorAuto, Value: "RECEBIDA"}},
	}

	repository.EXPECT().Listar(ctx, expectedParams).Return(query.Page[*domain.OrdemServico]{
		Items: []*domain.OrdemServico{os}, Total: 25, Page: 2, PageSize: 20, TotalPages: 2,
		Order: "data_cadastro", Direction: query.DirectionDESC,
	}, nil).Once()
	orcamentoRepository.EXPECT().BuscarPorOrdensServicoIDs(ctx, []uint64{42}).Return([]*domainorcamento.Orcamento{orcamento}, nil).Once()

	out, err := app.NewListarOrdensServicoUseCase(repository, orcamentoRepository).Executar(ctx, input)
	require.NoError(t, err)
	require.Len(t, out.Items, 1)
	require.Equal(t, int64(25), out.Total)
	require.Equal(t, 2, out.Page)
	require.Equal(t, 20, out.PageSize)
	require.Equal(t, 2, out.TotalPages)
	require.Equal(t, "data_cadastro", out.Order)
	require.Equal(t, "DESC", out.Direction)
	require.Equal(t, os.Numero().String(), out.Items[0].Numero)
	require.NotNil(t, out.Items[0].Orcamento)
	require.Equal(t, uint64(100), out.Items[0].Orcamento.ID)
	require.Equal(t, float64(450), out.Items[0].Orcamento.ValorTotal)
}

func TestListarOrdensServicoUseCase_Executar_ClienteForcaFiltroDoProprioID(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	orcamentoRepository := orcamentomocks.NewOrcamentoRepository(t)
	ctx := context.Background()

	expected := query.Params{
		Page: 1,
		Filters: []query.Filter{
			{Field: "status", Operator: query.OperatorAuto, Value: "RECEBIDA"},
			{Field: "cliente_id", Operator: query.OperatorEqual, Value: "10"},
		},
	}

	repository.EXPECT().Listar(ctx, expected).Return(query.Page[*domain.OrdemServico]{
		Items: []*domain.OrdemServico{}, Total: 0, Page: 1, PageSize: 20, TotalPages: 0,
		Order: "id", Direction: query.DirectionASC,
	}, nil).Once()

	out, err := app.NewListarOrdensServicoUseCase(repository, orcamentoRepository).Executar(ctx, app.ListarOrdensServicoInput{
		ParamsInput: appquery.ParamsInput{
			Page:    1,
			Filters: []appquery.FilterInput{{Field: "status", Operator: "auto", Value: "RECEBIDA"}},
		},
		SolicitanteID:   10,
		TipoSolicitante: domainauth.TipoCliente,
	})

	require.NoError(t, err)
	require.Empty(t, out.Items)
}

func TestListarOrdensServicoUseCase_Executar_ListaVazia(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	orcamentoRepository := orcamentomocks.NewOrcamentoRepository(t)
	repository.EXPECT().Listar(mock.Anything, query.Params{Page: 1}).Return(query.Page[*domain.OrdemServico]{
		Items: []*domain.OrdemServico{}, Total: 0, Page: 1, PageSize: 20, TotalPages: 0,
		Order: "id", Direction: query.DirectionASC,
	}, nil)

	out, err := app.NewListarOrdensServicoUseCase(repository, orcamentoRepository).Executar(context.Background(), app.ListarOrdensServicoInput{
		ParamsInput:     appquery.ParamsInput{Page: 1},
		SolicitanteID:   7,
		TipoSolicitante: domainauth.TipoInterno,
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
	orcamentoRepository := orcamentomocks.NewOrcamentoRepository(t)
	appErr := shared.NewValidationError("filtro inválido")
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(query.Page[*domain.OrdemServico]{}, appErr)

	_, err := app.NewListarOrdensServicoUseCase(repository, orcamentoRepository).Executar(context.Background(), app.ListarOrdensServicoInput{})
	require.ErrorIs(t, err, appErr)
}

func TestListarOrdensServicoUseCase_Executar_EncapsulaErroInfra(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	orcamentoRepository := orcamentomocks.NewOrcamentoRepository(t)
	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(query.Page[*domain.OrdemServico]{}, errors.New("banco indisponível"))

	_, err := app.NewListarOrdensServicoUseCase(repository, orcamentoRepository).Executar(context.Background(), app.ListarOrdensServicoInput{})
	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
}

func TestListarOrdensServicoUseCase_Executar_ErroAoBuscarOrcamentos_RetornaInternalError(t *testing.T) {
	repository := mocks.NewOrdemServicoRepository(t)
	orcamentoRepository := orcamentomocks.NewOrcamentoRepository(t)
	os := ordemServicoConsultaValida(t)

	repository.EXPECT().Listar(mock.Anything, mock.Anything).Return(query.Page[*domain.OrdemServico]{
		Items: []*domain.OrdemServico{os}, Total: 1, Page: 1, PageSize: 20, TotalPages: 1,
		Order: "id", Direction: query.DirectionASC,
	}, nil)
	orcamentoRepository.EXPECT().BuscarPorOrdensServicoIDs(mock.Anything, []uint64{42}).Return(nil, errors.New("banco indisponível"))

	_, err := app.NewListarOrdensServicoUseCase(repository, orcamentoRepository).Executar(context.Background(), app.ListarOrdensServicoInput{})
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

func orcamentoConsultaValido() *domainorcamento.Orcamento {
	agora := time.Date(2026, time.August, 30, 12, 30, 0, 0, time.UTC)
	return domainorcamento.ReidratarOrcamento(
		100,
		42,
		nil,
		nil,
		300,
		150,
		450,
		"Orçamento inicial",
		7,
		agora,
		agora,
	)
}
