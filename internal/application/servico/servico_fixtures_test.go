package servico_test

import (
	"testing"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/stretchr/testify/require"
)

func novoServico(t *testing.T) *domainservico.Servico {
	t.Helper()
	s, err := domainservico.NewServico("Troca de óleo", "Troca de óleo e filtro", 150.50, 60, 1)
	require.NoError(t, err)
	return s
}
