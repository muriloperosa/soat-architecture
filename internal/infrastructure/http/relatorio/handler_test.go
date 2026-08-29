package relatorio_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	app "github.com/muriloperosa/soat-architecture/internal/application/relatorio"
	domainrelatorio "github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
	relatoriomocks "github.com/muriloperosa/soat-architecture/internal/domain/relatorio/mocks"
	httprelatorio "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/relatorio"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func novoRouter(t *testing.T, repository *relatoriomocks.RelatorioTransicaoStatusRepository) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)
	handler := httprelatorio.NewHandler(app.NewConsultarTransicaoStatusUseCase(repository))
	router := gin.New()
	router.GET("/v1/relatorios/ordens-servico/transicao-status", handler.ConsultarTransicaoStatus)
	return router
}

func doGet(t *testing.T, router *gin.Engine, query string) (*httptest.ResponseRecorder, httprelatorio.TransicaoStatusResponse) {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/v1/relatorios/ordens-servico/transicao-status?"+query, nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	var response httprelatorio.TransicaoStatusResponse
	if rec.Body.Len() > 0 {
		_ = json.Unmarshal(rec.Body.Bytes(), &response)
	}
	return rec, response
}

func TestHandlerConsultarTransicaoStatus_ComSucesso_UnidadePadraoHoras(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	repository.EXPECT().
		CalcularTransicaoStatus(mock.Anything, mock.Anything).
		Return(domainrelatorio.TransicaoStatusResultado{
			TotalOrdens:   2,
			DuracaoMedia:  90 * time.Minute,
			DuracaoMinima: 30 * time.Minute,
			DuracaoMaxima: 150 * time.Minute,
		}, nil)

	router := novoRouter(t, repository)
	rec, response := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=RECEBIDA&to_status=ENTREGUE")

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, 2, response.TotalOrdensServico)
	require.Equal(t, "h", response.Unidade)
	require.InDelta(t, 1.5, response.TempoMedio, 0.0001)
	require.InDelta(t, 0.5, response.TempoMinimo, 0.0001)
	require.InDelta(t, 2.5, response.TempoMaximo, 0.0001)
}

func TestHandlerConsultarTransicaoStatus_UnidadeMinutos(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	repository.EXPECT().
		CalcularTransicaoStatus(mock.Anything, mock.Anything).
		Return(domainrelatorio.TransicaoStatusResultado{TotalOrdens: 1, DuracaoMedia: 2 * time.Hour}, nil)

	router := novoRouter(t, repository)
	rec, response := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=RECEBIDA&to_status=ENTREGUE&unit=m")

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "m", response.Unidade)
	require.InDelta(t, 120, response.TempoMedio, 0.0001)
}

func TestHandlerConsultarTransicaoStatus_UnidadeSegundos(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	repository.EXPECT().
		CalcularTransicaoStatus(mock.Anything, mock.Anything).
		Return(domainrelatorio.TransicaoStatusResultado{TotalOrdens: 1, DuracaoMedia: 2 * time.Minute}, nil)

	router := novoRouter(t, repository)
	rec, response := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=RECEBIDA&to_status=ENTREGUE&unit=s")

	require.Equal(t, http.StatusOK, rec.Code)
	require.Equal(t, "s", response.Unidade)
	require.InDelta(t, 120, response.TempoMedio, 0.0001)
}

func TestHandlerConsultarTransicaoStatus_UnidadeInvalida_Retorna400(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	router := novoRouter(t, repository)

	rec, _ := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=RECEBIDA&to_status=ENTREGUE&unit=dias")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerConsultarTransicaoStatus_QueryParamsAusentes_Retorna400(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	router := novoRouter(t, repository)

	rec, _ := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=RECEBIDA")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerConsultarTransicaoStatus_StartDateFormatoInvalido_Retorna400(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	router := novoRouter(t, repository)

	rec, _ := doGet(t, router, "start_date=01-01-2026&final_date=2026-08-28&from_status=RECEBIDA&to_status=ENTREGUE")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerConsultarTransicaoStatus_FinalDateFormatoInvalido_Retorna400(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	router := novoRouter(t, repository)

	rec, _ := doGet(t, router, "start_date=2026-01-01&final_date=28-08-2026&from_status=RECEBIDA&to_status=ENTREGUE")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerConsultarTransicaoStatus_StatusInvalido_Retorna400(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	router := novoRouter(t, repository)

	rec, _ := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=INVALIDO&to_status=ENTREGUE")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerConsultarTransicaoStatus_SemCaminhoValido_Retorna400(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	router := novoRouter(t, repository)

	rec, _ := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=APROVADA&to_status=RECEBIDA")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerConsultarTransicaoStatus_PeriodoInvalido_Retorna400(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	router := novoRouter(t, repository)

	rec, _ := doGet(t, router, "start_date=2026-08-28&final_date=2026-01-01&from_status=RECEBIDA&to_status=ENTREGUE")

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestHandlerConsultarTransicaoStatus_ErroInterno_Retorna500(t *testing.T) {
	repository := relatoriomocks.NewRelatorioTransicaoStatusRepository(t)
	repository.EXPECT().
		CalcularTransicaoStatus(mock.Anything, mock.Anything).
		Return(domainrelatorio.TransicaoStatusResultado{}, errors.New("erro de conexão"))

	router := novoRouter(t, repository)
	rec, _ := doGet(t, router, "start_date=2026-01-01&final_date=2026-08-28&from_status=RECEBIDA&to_status=ENTREGUE")

	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
