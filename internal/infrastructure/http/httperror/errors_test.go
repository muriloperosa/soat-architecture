package httperror_test

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
	"github.com/stretchr/testify/require"
)

func serveRespondError(err error) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/x", func(c *gin.Context) {
		httperror.RespondError(c, err)
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

func TestRespondError_Forbidden_Retorna403(t *testing.T) {
	rec := serveRespondError(shared.NewForbiddenError("acesso não permitido para este tipo de usuário"))

	require.Equal(t, http.StatusForbidden, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "forbidden", body.Type)
}

func TestRespondError_Unauthorized_Retorna401(t *testing.T) {
	rec := serveRespondError(shared.NewUnauthorizedError("token inválido ou expirado"))

	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "unauthorized", body.Type)
}

func TestRespondError_Internal_Retorna500SemVazarDetalheInterno(t *testing.T) {
	rec := serveRespondError(shared.NewInternalError("erro ao consultar banco", errors.New("connection refused")))

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "internal", body.Type)
	require.NotContains(t, body.Message, "connection refused")
}

func TestRespondError_Unavailable_Retorna503SemVazarDetalheInterno(t *testing.T) {
	rec := serveRespondError(shared.NewUnavailableError("banco de dados indisponível", errors.New("connection refused")))

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "unavailable", body.Type)
	require.Equal(t, "banco de dados indisponível", body.Message)
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

func serveRespond(fn func(c *gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.GET("/x", fn)

	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	return rec
}

func TestRespondValidationError_Retorna400(t *testing.T) {
	rec := serveRespond(func(c *gin.Context) {
		httperror.RespondValidationError(c, "corpo inválido")
	})

	require.Equal(t, http.StatusBadRequest, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "validation", body.Type)
	require.Equal(t, "corpo inválido", body.Message)
}

func TestRespondNotFoundError_Retorna404(t *testing.T) {
	rec := serveRespond(func(c *gin.Context) {
		httperror.RespondNotFoundError(c, "cliente não encontrado")
	})

	require.Equal(t, http.StatusNotFound, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "not_found", body.Type)
	require.Equal(t, "cliente não encontrado", body.Message)
}

func TestRespondConflictError_Retorna409(t *testing.T) {
	rec := serveRespond(func(c *gin.Context) {
		httperror.RespondConflictError(c, "ordem já finalizada")
	})

	require.Equal(t, http.StatusConflict, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "conflict", body.Type)
}

func TestRespondForbiddenError_Retorna403(t *testing.T) {
	rec := serveRespond(func(c *gin.Context) {
		httperror.RespondForbiddenError(c, "acesso não permitido")
	})

	require.Equal(t, http.StatusForbidden, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "forbidden", body.Type)
}

func TestRespondUnauthorizedError_Retorna401(t *testing.T) {
	rec := serveRespond(func(c *gin.Context) {
		httperror.RespondUnauthorizedError(c, "token inválido")
	})

	require.Equal(t, http.StatusUnauthorized, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "unauthorized", body.Type)
}

func TestRespondInternalError_Retorna500SemVazarDetalheInterno(t *testing.T) {
	rec := serveRespond(func(c *gin.Context) {
		httperror.RespondInternalError(c, "erro ao consultar banco", errors.New("connection refused"))
	})

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "internal", body.Type)
	require.NotContains(t, body.Message, "connection refused")
}

func TestRespondUnavailableError_Retorna503(t *testing.T) {
	rec := serveRespond(func(c *gin.Context) {
		httperror.RespondUnavailableError(c, "banco de dados indisponível", errors.New("connection refused"))
	})

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)

	var body errorBody
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	require.Equal(t, "unavailable", body.Type)
	require.Equal(t, "banco de dados indisponível", body.Message)
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
