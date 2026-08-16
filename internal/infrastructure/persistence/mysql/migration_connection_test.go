package mysql

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
)

func TestBuildMigrationDSN(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "3306",
		DBUser:     "root",
		DBPassword: "root",
		DBName:     "mecanica",
	}

	dsn := buildMigrationDSN(cfg)

	require.Equal(
		t,
		"root:root@tcp(127.0.0.1:3306)/mecanica?charset=utf8mb4&parseTime=True&loc=Local&timeout=10s&readTimeout=30s&writeTimeout=30s",
		dsn,
	)
}

func TestConfigureMigrationPool(t *testing.T) {
	sqlDB, _, err := sqlmock.New()
	require.NoError(t, err)
	defer sqlDB.Close()

	configureMigrationPool(sqlDB)

	require.Equal(t, 1, sqlDB.Stats().MaxOpenConnections)
}

func TestNewMigrationConnectionComConfigNil(t *testing.T) {
	db, err := NewMigrationConnection(nil)

	require.Nil(t, db)
	require.Error(t, err)
	require.Equal(t, ErrConfigNotFound, err)
}

func TestNewMigrationConnectionComBancoIndisponivel(t *testing.T) {
	cfg := &config.Config{
		DBHost:     "127.0.0.1",
		DBPort:     "1",
		DBUser:     "root",
		DBPassword: "root",
		DBName:     "mecanica",
	}

	db, err := NewMigrationConnection(cfg)

	require.Nil(t, db)
	require.Error(t, err)
}
