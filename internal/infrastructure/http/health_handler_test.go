package http_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	handler "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
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

	router := handler.NewRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)
	require.NoError(t, mock.ExpectationsWereMet())
}

func TestHealthHandler_Retorna503QuandoBancoFalha(t *testing.T) {
	db, mock := newMockedGormDB(t)
	mock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)

	router := handler.NewRouter(db)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	require.Equal(t, http.StatusServiceUnavailable, rec.Code)
}
