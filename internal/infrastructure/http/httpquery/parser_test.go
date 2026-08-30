package httpquery

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func testContext(target string) *gin.Context {
	gin.SetMode(gin.TestMode)
	rec := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(rec)
	ctx.Request = httptest.NewRequest("GET", target, nil)
	return ctx
}

func TestParser_Parse_ComPaginacaoOrdenacaoEFiltros(t *testing.T) {
	params, err := NewParser().Parse(testContext("/v1/clientes?page=3&order=nome&direction=desc&nome=Maria&ativo=true"))

	require.NoError(t, err)
	require.Equal(t, 3, params.Page)
	require.Equal(t, "nome", params.Order)
	require.Equal(t, "DESC", params.Direction)
	require.Len(t, params.Filters, 2)
	require.Equal(t, "ativo", params.Filters[0].Field)
	require.Equal(t, OperatorAuto, params.Filters[0].Operator)
	require.Equal(t, "nome", params.Filters[1].Field)
}

func TestParser_Parse_SemPaginacao_UsaPrimeiraPagina(t *testing.T) {
	params, err := NewParser().Parse(testContext("/v1/clientes"))

	require.NoError(t, err)
	require.Equal(t, 1, params.Page)
	require.Empty(t, params.Order)
	require.Empty(t, params.Direction)
	require.NotNil(t, params.Filters)
}

func TestParser_Parse_PageInvalida_RetornaErro(t *testing.T) {
	_, err := NewParser().Parse(testContext("/v1/clientes?page=invalida"))
	require.Error(t, err)
}

func TestParser_Parse_PageZero_RetornaErro(t *testing.T) {
	_, err := NewParser().Parse(testContext("/v1/clientes?page=0"))
	require.Error(t, err)
}

func TestParser_Parse_FiltroSemValor_RetornaErro(t *testing.T) {
	_, err := NewParser().Parse(testContext("/v1/clientes?nome="))
	require.Error(t, err)
}

func TestParser_Parse_FiltroNegado(t *testing.T) {
	params, err := NewParser().Parse(testContext("/v1/clientes?email_not=teste@email.com"))

	require.NoError(t, err)
	require.Len(t, params.Filters, 1)
	require.Equal(t, "email", params.Filters[0].Field)
	require.Equal(t, OperatorAutoNot, params.Filters[0].Operator)
	require.Equal(t, "teste@email.com", params.Filters[0].Value)
}
