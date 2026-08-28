package httpquery

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func testContext(target string) *gin.Context {
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(http.MethodGet, target, nil)
	return context
}

func TestParserParse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	parser := NewParser()

	params, err := parser.Parse(testContext(
		"/v1/clientes?offset=10&limit=5&order=nome&direction=desc" +
			"&nome=Maria&ativo_not=false&id=1,2,3",
	))

	require.NoError(t, err)
	require.Equal(t, 10, params.Offset)
	require.Equal(t, 5, params.Limit)
	require.Equal(t, "nome", params.Order)
	require.Equal(t, query.DirectionDESC, params.Direction)
	require.Equal(t, []query.Filter{
		{Field: "ativo", Operator: query.OperatorAutoNot, Value: "false"},
		{Field: "id", Operator: query.OperatorAuto, Value: "1,2,3"},
		{Field: "nome", Operator: query.OperatorAuto, Value: "Maria"},
	}, params.Filters)
}

func TestParserParseMantemPadroesParaParametrosOmitidos(t *testing.T) {
	gin.SetMode(gin.TestMode)

	params, err := NewParser().Parse(testContext("/v1/clientes"))

	require.NoError(t, err)
	require.Zero(t, params.Offset)
	require.Zero(t, params.Limit)
	require.Empty(t, params.Order)
	require.Empty(t, params.Direction)
	require.Empty(t, params.Filters)
}

func TestParserParseRejeitaNumeroInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, err := NewParser().Parse(testContext("/v1/clientes?limit=invalido"))

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestParserParseRejeitaFiltroSemValor(t *testing.T) {
	gin.SetMode(gin.TestMode)

	_, err := NewParser().Parse(testContext("/v1/clientes?nome="))

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindValidation, appErr.Kind)
}

func TestParserParseReconheceSufixoDeNegacao(t *testing.T) {
	gin.SetMode(gin.TestMode)

	params, err := NewParser().Parse(testContext("/v1/clientes?email_not=teste@email.com"))

	require.NoError(t, err)
	require.Equal(t, []query.Filter{{
		Field: "email", Operator: query.OperatorAutoNot, Value: "teste@email.com",
	}}, params.Filters)
}
