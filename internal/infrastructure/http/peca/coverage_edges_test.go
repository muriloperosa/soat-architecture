package peca_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	httpp "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/peca"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandlerPeca_AtualizarEInativarRejeitamIDInvalido(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := httpp.NewHandler(nil, nil, nil, nil, nil, nil, nil, nil)
	for _, tt := range []struct {
		name, method, route string
		handler             gin.HandlerFunc
	}{
		{"atualizar", http.MethodPut, "/invalid", h.Atualizar},
		{"inativar", http.MethodPatch, "/invalid", h.Inativar},
	} {
		t.Run(tt.name, func(t *testing.T) {
			router := gin.New()
			router.Handle(tt.method, "/:id", tt.handler)
			rec := httptest.NewRecorder()
			router.ServeHTTP(rec, httptest.NewRequest(tt.method, tt.route, nil))
			require.Equal(t, http.StatusBadRequest, rec.Code)
		})
	}
}

func TestHandlerPeca_InativarPropagaErroDoCasoDeUso(t *testing.T) {
	gin.SetMode(gin.TestMode)
	repo := mocks.NewRepository(t)
	errConsulta := errors.New("falha de persistência")
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(nil, errConsulta)
	h := httpp.NewHandler(nil, nil, nil, apppeca.NewInativarPecaUseCase(repo), nil, nil, nil, nil)
	router := gin.New()
	router.PATCH("/:id/inativar", h.Inativar)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, httptest.NewRequest(http.MethodPatch, "/1/inativar", nil))
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}
