package httprequest_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httprequest"
	"github.com/stretchr/testify/require"
)

func TestParseDateQueryParam_Valido_RetornaValorEmUTC(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/recurso", func(c *gin.Context) {
		data, ok := httprequest.ParseDateQueryParam(c, "data")
		require.True(t, ok)
		require.True(t, data.Equal(time.Date(2026, time.March, 15, 0, 0, 0, 0, time.UTC)))
		require.Equal(t, time.UTC, data.Location())
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/recurso?data=2026-03-15", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
}

func TestParseDateQueryParam_Ausente_Retorna400(t *testing.T) {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/recurso", func(c *gin.Context) {
		_, ok := httprequest.ParseDateQueryParam(c, "data")
		require.False(t, ok)
	})

	req := httptest.NewRequest(http.MethodGet, "/recurso", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func TestParseDateQueryParam_FormatoInvalido_Retorna400(t *testing.T) {
	testes := []string{"15-03-2026", "2026/03/15", "não é data", ""}

	for _, valor := range testes {
		t.Run(valor, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			engine := gin.New()
			engine.GET("/recurso", func(c *gin.Context) {
				_, ok := httprequest.ParseDateQueryParam(c, "data")
				require.False(t, ok)
			})

			req := httptest.NewRequest(http.MethodGet, "/recurso?data="+url.QueryEscape(valor), nil)
			rec := httptest.NewRecorder()
			engine.ServeHTTP(rec, req)

			require.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}
