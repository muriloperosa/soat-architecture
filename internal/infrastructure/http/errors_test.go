package http_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	handler "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
	"github.com/stretchr/testify/require"
)

func serveRespondError(err error) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/x", func(c *gin.Context) {
		handler.RespondError(c, err)
	})

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

type errorBody struct {
	Type    string   `json:"type"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}

func TestRespondError_NotFound_Retorna404(t *testing.T) {
	rec := serveRespondError(shared.NewNotFoundError("cliente não encontrado"))

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "not_found", body.Type)
	require.Equal(t, "cliente não encontrado", body.Message)
}

func TestRespondError_Validation_Retorna400ComDetails(t *testing.T) {
	rec := serveRespondError(shared.NewValidationErrorWithDetails("dados inválidos", []string{"nome é obrigatório"}))

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "validation", body.Type)
	require.Equal(t, []string{"nome é obrigatório"}, body.Details)
}

func TestRespondError_Conflict_Retorna409(t *testing.T) {
	rec := serveRespondError(shared.NewConflictError("ordem já finalizada"))

	require.Equal(t, http.StatusConflict, rec.Code)
}

func TestRespondError_Internal_Retorna500SemVazarDetalheInterno(t *testing.T) {
	rec := serveRespondError(shared.NewInternalError("erro ao consultar banco", errors.New("connection refused")))

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "internal", body.Type)
	require.NotContains(t, body.Message, "connection refused")
}

func TestRespondError_ErroDesconhecido_Retorna500Generico(t *testing.T) {
	rec := serveRespondError(errors.New("algo inesperado"))

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "internal", body.Type)
	require.NotContains(t, body.Message, "algo inesperado")
}

func TestRespondError_ErroComKindInvalido_Retorna500Generico(t *testing.T) {
	invalidKind := shared.ErrorKind("invalid_kind")

	err := shared.AppError{
		Kind:    invalidKind,
		Err:     errors.New("erro desconhecido"),
		Message: "erro desconhecido",
		Details: []string{"erro desconhecido"},
	}

	rec := serveRespondError(&err)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "invalid_kind", body.Type)
	require.Equal(t, "erro desconhecido", body.Message)
	require.Equal(t, []string{"erro desconhecido"}, body.Details)
	require.Equal(t, rec.Code, http.StatusInternalServerError)
}
