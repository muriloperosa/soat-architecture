package mysql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/mysql"
)

func TestNewConnection_ErroComHostInvalido(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "host-inexistente",
		DBPort:     "3306",
		DBUser:     "root",
		DBPassword: "root",
		DBName:     "mecanica",
	}

	_, err := NewConnection(cfg)

	require.Error(t, err)
}

func TestOpenConnection_ConfiguraPool(t *testing.T) {
	sqlDB, sqlMock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	require.NoError(t, err)
	defer sqlDB.Close()
	sqlMock.ExpectPing()

	cfg := &config.Config{DBMaxOpenConns: 50, DBMaxIdleConns: 10, DBConnMaxLifetimeMinutes: 15}
	db, err := openConnection(mysql.New(mysql.Config{Conn: sqlDB, SkipInitializeWithVersion: true}), cfg)

	require.NoError(t, err)
	underlyingDB, err := db.DB()
	require.NoError(t, err)
	require.Equal(t, 50, underlyingDB.Stats().MaxOpenConnections)
	require.NoError(t, sqlMock.ExpectationsWereMet())
}

func TestConfigurePool_AplicaLimitesDoConfig(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	cfg := &config.Config{
		DBMaxOpenConns:           50,
		DBMaxIdleConns:           10,
		DBConnMaxLifetimeMinutes: 15,
	}

	configurePool(sqlDB, cfg)

	stats := sqlDB.Stats()
	require.Equal(t, 50, stats.MaxOpenConnections)
}
