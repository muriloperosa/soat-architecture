package mysql_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
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

	_, err := mysql.NewConnection(cfg)

	require.Error(t, err)
}
