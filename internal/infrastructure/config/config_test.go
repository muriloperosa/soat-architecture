package config_test

import (
	"os"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_UsaValoresDasEnvs(t *testing.T) {
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "9090", cfg.AppPort)
	require.Equal(t, "db-host", cfg.DBHost)
	require.Equal(t, "3307", cfg.DBPort)
	require.Equal(t, "user", cfg.DBUser)
	require.Equal(t, "pass", cfg.DBPassword)
	require.Equal(t, "oficina_test", cfg.DBName)
}

func TestLoad_UsaDefaultDeAppPort(t *testing.T) {
	os.Unsetenv("APP_PORT")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "8080", cfg.AppPort)
}
