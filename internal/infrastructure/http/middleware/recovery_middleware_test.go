package middleware_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/middleware"
	"github.com/stretchr/testify/require"
)

type recoveryErrorBody struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func setupEngineComRecovery(panicValue any) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middleware.RecoveryMiddleware())
	engine.GET("/quebra", func(c *gin.Context) {
		panic(panicValue)
	})
	return engine
}

func TestRecoveryMiddleware_PanicComError_Retorna500SemVazarErroOriginal(t *testing.T) {
	engine := setupEngineComRecovery(errors.New("connection refused"))

	req := httptest.NewRequest(http.MethodGet, "/quebra", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body recoveryErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "internal", body.Type)
	require.Equal(t, "erro interno", body.Message)
	require.NotContains(t, rec.Body.String(), "connection refused")
}

func TestRecoveryMiddleware_PanicComValorNaoError_Retorna500(t *testing.T) {
	engine := setupEngineComRecovery("algo quebrou")

	req := httptest.NewRequest(http.MethodGet, "/quebra", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body recoveryErrorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "internal", body.Type)
	require.Equal(t, "erro interno", body.Message)
}

func TestRecoveryMiddleware_NaoVazaTextoPlanoDeErroPadraoDoGin(t *testing.T) {
	engine := setupEngineComRecovery(errors.New("boom"))

	req := httptest.NewRequest(http.MethodGet, "/quebra", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)

	require.Equal(t, "application/json; charset=utf-8", rec.Header().Get("Content-Type"))
}
