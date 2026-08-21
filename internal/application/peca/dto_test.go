package peca

import (
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/stretchr/testify/require"
)

func TestToOutput(t *testing.T) {
	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	p.AtribuirID(10)

	output := toOutput(p)

	require.Equal(t, uint64(10), output.ID)
	require.Equal(t, p.Codigo(), output.Codigo)
	require.Equal(t, "Peca 1", output.Nome)
	require.Equal(t, "Marca 1", output.Marca)
	require.Equal(t, "Descricao 1", output.Descricao)
	require.Equal(t, 100.0, output.Preco)
	require.Equal(t, 10, output.QuantidadeEmEstoque)
	require.Equal(t, 5, output.EstoqueMinimo)
	require.Equal(t, uint64(1), output.CriadoPor)
	require.True(t, output.Ativo)
}
