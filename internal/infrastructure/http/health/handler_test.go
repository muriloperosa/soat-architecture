package health

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func TestHealthHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("falha ao obter conexão sql", func(t *testing.T) {
		pool := failingDBConnector{}
		db := &gorm.DB{Config: &gorm.Config{ConnPool: pool}}
		router := gin.New()
		router.GET("/health", NewHealthHandler(db))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))
		require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
	})

	t.Run("banco responde", func(t *testing.T) {
		sqlDB, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer sqlDB.Close()
		db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
		require.NoError(t, err)
		sqlMock.ExpectPing()

		router := gin.New()
		router.GET("/health", NewHealthHandler(db))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

		require.Equal(t, http.StatusOK, recorder.Code)
		require.NoError(t, sqlMock.ExpectationsWereMet())
	})

	t.Run("falha no ping", func(t *testing.T) {
		sqlDB, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
		require.NoError(t, err)
		defer sqlDB.Close()
		db, err := gorm.Open(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), &gorm.Config{DisableAutomaticPing: true})
		require.NoError(t, err)
		sqlMock.ExpectPing().WillReturnError(sqlmock.ErrCancelled)

		router := gin.New()
		router.GET("/health", NewHealthHandler(db))
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/health", nil))

		require.Equal(t, http.StatusServiceUnavailable, recorder.Code)
		require.NoError(t, sqlMock.ExpectationsWereMet())
	})
}

type failingDBConnector struct{}

func (failingDBConnector) PrepareContext(context.Context, string) (*sql.Stmt, error) {
	return nil, errors.New("not implemented")
}
func (failingDBConnector) ExecContext(context.Context, string, ...interface{}) (sql.Result, error) {
	return nil, errors.New("not implemented")
}
func (failingDBConnector) QueryContext(context.Context, string, ...interface{}) (*sql.Rows, error) {
	return nil, errors.New("not implemented")
}
func (failingDBConnector) QueryRowContext(context.Context, string, ...interface{}) *sql.Row {
	return nil
}
func (failingDBConnector) GetDBConn() (*sql.DB, error) {
	return nil, errors.New("connection unavailable")
}
