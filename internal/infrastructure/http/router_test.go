package http_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	handler "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func newMockedGormDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()

	sqlDB, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)

	gormDB, err := gorm.Open(mysql.New(mysql.Config{
		Conn:                      sqlDB,
		SkipInitializeWithVersion: true,
	}), &gorm.Config{DisableAutomaticPing: true})
	require.NoError(t, err)

	return gormDB, mock
}

func TestHealthHandler_RetornaOKQuandoBancoResponde(t *testing.T) {
	db, mock := newMockedGormDB(t)
	mock.ExpectPing()

	container, err := wiring.NewContainer(nil, db)
	require.NoError(t, err)
	router := handler.NewRouter(container)

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthHandler_Retorna503QuandoBancoFalha(t *testing.T) {
	db, mock := newMockedGormDB(t)
	mock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)

	container, err := wiring.NewContainer(nil, db)
	require.NoError(t, err)
	router := handler.NewRouter(container)

	req := httptest.NewRequest(http.MethodGet, "/v1/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestHealthLive_RetornaOKSemTocarNoBanco(t *testing.T) {
	db, mock := newMockedGormDB(t)

	container, err := wiring.NewContainer(nil, db)
	require.NoError(t, err)
	router := handler.NewRouter(container)

	req := httptest.NewRequest(http.MethodGet, "/v1/health/live", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.JSONEq(t, `{"status":"ok"}`, rec.Body.String())
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthReady_RetornaOKQuandoBancoResponde(t *testing.T) {
	db, mock := newMockedGormDB(t)
	mock.ExpectPing()

	container, err := wiring.NewContainer(nil, db)
	require.NoError(t, err)
	router := handler.NewRouter(container)

	req := httptest.NewRequest(http.MethodGet, "/v1/health/ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthReady_Retorna503QuandoBancoFalha(t *testing.T) {
	db, mock := newMockedGormDB(t)
	mock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)

	container, err := wiring.NewContainer(nil, db)
	require.NoError(t, err)
	router := handler.NewRouter(container)

	req := httptest.NewRequest(http.MethodGet, "/v1/health/ready", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}

func TestRouter_NaoRegistraTokenDeAprovacaoNoAccessLog(t *testing.T) {
	previousWriter := gin.DefaultWriter
	var logOutput bytes.Buffer
	gin.DefaultWriter = &logOutput
	t.Cleanup(func() { gin.DefaultWriter = previousWriter })

	db, _ := newMockedGormDB(t)
	container, err := wiring.NewContainer(nil, db)
	require.NoError(t, err)
	router := handler.NewRouter(container)

	secretToken := "eyJhbGciOiJIUzI1NiJ9.secret.signature"
	request := httptest.NewRequest(http.MethodGet, "/v1/orcamentos/aprovacao?token="+secretToken+"&acao=APROVAR", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	require.NotContains(t, logOutput.String(), secretToken)

	request = httptest.NewRequest(http.MethodGet, "/v1/health/live", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	require.Contains(t, logOutput.String(), "/v1/health/live", "o access log continua ativo para outras rotas")
	require.NotContains(t, logOutput.String(), secretToken)
}
