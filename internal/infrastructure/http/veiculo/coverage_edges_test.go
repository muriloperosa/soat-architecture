package veiculo_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	httpv "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/veiculo"
	"github.com/stretchr/testify/require"
)

func TestHandlerVeiculo_AtualizarRejeitaIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := httpv.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	router := gin.New()
	router.PUT("/:id", h.Atualizar)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPut, "/invalid", nil))
	require.Equal(t, http.StatusBadRequest, rec.Code)
}
