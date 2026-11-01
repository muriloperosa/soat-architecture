package config_test

import (
	"os"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/stretchr/testify/require"
)

func TestLoad_UsaValoresDasEnvs(t *testing.T) {
	unsetEnv(t, "EMAIL_PROVIDER", "AWS_REGION", "SES_FROM_ADDRESS", "APPROVAL_TOKEN_SECRET", "APPROVAL_TOKEN_TTL_HOURS", "PUBLIC_BASE_URL")
	t.Setenv("APP_PORT", "9090")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3307")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "9090", cfg.AppPort)
	require.Equal(t, "db-host", cfg.DBHost)
	require.Equal(t, "3307", cfg.DBPort)
	require.Equal(t, "user", cfg.DBUser)
	require.Equal(t, "pass", cfg.DBPassword)
	require.Equal(t, "oficina_test", cfg.DBName)
	require.Equal(t, "log", cfg.EmailProvider)
	require.Equal(t, "us-east-1", cfg.AWSRegion)
	require.Empty(t, cfg.SESFromAddress)
	require.Empty(t, cfg.ApprovalTokenSecret)
	require.Equal(t, 24, cfg.ApprovalTokenTTLHours)
}

func unsetEnv(t *testing.T, names ...string) {
	t.Helper()
	for _, name := range names {
		name := name
		previous, existed := os.LookupEnv(name)
		require.NoError(t, os.Unsetenv(name))
		t.Cleanup(func() {
			if existed {
				require.NoError(t, os.Setenv(name, previous))
			} else {
				require.NoError(t, os.Unsetenv(name))
			}
		})
	}
}

func TestLoad_UsaValoresDoPoolDasEnvs(t *testing.T) {
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")
	t.Setenv("DB_MAX_OPEN_CONNS", "50")
	t.Setenv("DB_MAX_IDLE_CONNS", "10")
	t.Setenv("DB_CONN_MAX_LIFETIME_MINUTES", "15")
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, 50, cfg.DBMaxOpenConns)
	require.Equal(t, 10, cfg.DBMaxIdleConns)
	require.Equal(t, 15, cfg.DBConnMaxLifetimeMinutes)
}

func TestLoad_UsaDefaultsDoPoolQuandoEnvsNaoDefinidas(t *testing.T) {
	os.Unsetenv("DB_MAX_OPEN_CONNS")
	os.Unsetenv("DB_MAX_IDLE_CONNS")
	os.Unsetenv("DB_CONN_MAX_LIFETIME_MINUTES")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, 25, cfg.DBMaxOpenConns)
	require.Equal(t, 5, cfg.DBMaxIdleConns)
	require.Equal(t, 5, cfg.DBConnMaxLifetimeMinutes)
}

func TestLoad_UsaDefaultDeAppPort(t *testing.T) {
	os.Unsetenv("APP_PORT")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "8080", cfg.AppPort)
}

func TestLoad_ErroQuandoEnvObrigatoriaFaltando(t *testing.T) {
	os.Unsetenv("DB_HOST")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	cfg, err := config.Load()

	require.Error(t, err)
	require.Nil(t, cfg)
}

func TestLoad_CamposJWT(t *testing.T) {
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")
	t.Setenv("JWT_SECRET", "segredo-de-teste")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "segredo-de-teste", cfg.JWTSecret)
	require.Equal(t, 60, cfg.JWTAccessTokenTTLMinutes)
	require.Equal(t, 168, cfg.JWTRefreshTokenTTLHours)
}

func TestLoad_CamposTokenAprovacao(t *testing.T) {
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")
	t.Setenv("JWT_SECRET", "segredo-de-teste")
	t.Setenv("APPROVAL_TOKEN_SECRET", "segredo-de-aprovacao")
	t.Setenv("APPROVAL_TOKEN_TTL_HOURS", "12")
	t.Setenv("PUBLIC_BASE_URL", "https://oficina.example")

	cfg, err := config.Load()

	require.NoError(t, err)
	require.Equal(t, "segredo-de-aprovacao", cfg.ApprovalTokenSecret)
	require.Equal(t, 12, cfg.ApprovalTokenTTLHours)
	require.Equal(t, "https://oficina.example", cfg.PublicBaseURL)
}

func TestLoad_JWTSecretObrigatorio(t *testing.T) {
	os.Unsetenv("JWT_SECRET")
	t.Setenv("DB_HOST", "db-host")
	t.Setenv("DB_PORT", "3306")
	t.Setenv("DB_USER", "user")
	t.Setenv("DB_PASSWORD", "pass")
	t.Setenv("DB_NAME", "oficina_test")

	cfg, err := config.Load()

	require.Error(t, err)
	require.Nil(t, cfg)
}
