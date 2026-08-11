package mysql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
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
	require.Equal(t, 60, stats.MaxOpenConnections)
}
