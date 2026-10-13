package servico_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	https "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/servico"
	"github.com/stretchr/testify/require"
)

func TestHandlerServico_AtualizarRejeitaIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := https.NewHandler(nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.PUT("/:id", h.Atualizar)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/invalid", nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
