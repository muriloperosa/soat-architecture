package database

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDriverTypeString(t *testing.T) {
	tests := []struct {
		name     string
		driver   DriverType
		expected string
	}{
		{
			name:     "mysql",
			driver:   DriverTypeMySQL,
			expected: "mysql",
		},
		{
			name:     "postgresql",
			driver:   DriverTypePostgreSQL,
			expected: "postgres",
		},
		{
			name:     "sqlite",
			driver:   DriverTypeSQLite,
			expected: "sqlite",
		},
		{
			name:     "sql server",
			driver:   DriverTypeSQLServer,
			expected: "sqlserver",
		},
		{
			name:     "oracle",
			driver:   DriverTypeOracle,
			expected: "oracle",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.driver.String())
		})
	}
}

func TestDriverTypeIsValid(t *testing.T) {
	tests := []struct {
		name     string
		driver   DriverType
		expected bool
	}{
		{
			name:     "mysql válido",
			driver:   DriverTypeMySQL,
			expected: true,
		},
		{
			name:     "postgresql válido",
			driver:   DriverTypePostgreSQL,
			expected: true,
		},
		{
			name:     "sqlite válido",
			driver:   DriverTypeSQLite,
			expected: true,
		},
		{
			name:     "sql server válido",
			driver:   DriverTypeSQLServer,
			expected: true,
		},
		{
			name:     "oracle válido",
			driver:   DriverTypeOracle,
			expected: true,
		},
		{
			name:     "driver vazio",
			driver:   DriverType(""),
			expected: false,
		},
		{
			name:     "driver desconhecido",
			driver:   DriverType("mongodb"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, tt.driver.IsValid())
		})
	}
}
