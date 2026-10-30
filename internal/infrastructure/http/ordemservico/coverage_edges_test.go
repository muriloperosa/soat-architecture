package ordemservico_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	httpos "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
	"github.com/stretchr/testify/require"
)

func TestHandlerOrdemServico_CobreRamosDeEntradaInvalida(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := httpos.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
	t.Run("listar sem claims", func(t *testing.T) {
		router := gin.New()
		router.GET("/ordens", h.Listar)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/ordens", nil))
		require.Equal(t, http.StatusUnauthorized, rec.Code)
	})
	t.Run("abrir com body inválido", func(t *testing.T) {
		router := gin.New()
		router.Use(comClaimsOrdemServico)
		router.POST("/ordens", h.Abrir)
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/ordens", nil)
		router.ServeHTTP(rec, req)
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
	t.Run("diagnóstico com ID inválido", func(t *testing.T) {
		router := gin.New()
		router.PUT("/:id/diagnostico", h.InformarDiagnostico)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/invalid/diagnostico", nil))
		require.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func comClaimsOrdemServico(c *gin.Context) {
	c.Set(middleware.ClaimsContextKey, &domainauth.AppClaims{Subject: "1", Tipo: domainauth.TipoInterno, Papel: shared.PapelAdmin})
	c.Next()
}
