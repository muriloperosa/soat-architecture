package httprequest_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/stretchr/testify/require"
)

func TestParseUintParam_Valido_RetornaValor(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/recurso/:id", func(c *gin.Context) {
		id, ok := httprequest.ParseUintParam(c, "id")
		require.True(t, ok)
		c.JSON(http.StatusOK, gin.H{"id": id})
	})

	req := httptest.NewRequest(http.MethodGet, "/recurso/42", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestParseUintParam_Invalido_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/recurso/:id", func(c *gin.Context) {
		_, ok := httprequest.ParseUintParam(c, "id")
		require.False(t, ok)
	})

	req := httptest.NewRequest(http.MethodGet, "/recurso/nao-e-numero", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}
