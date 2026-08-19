package cliente

import (
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/stretchr/testify/require"
)

func TestToOutput(t *testing.T) {
	cliente, err := domain.NewCliente(
		"529.982.247-25",
		domain.TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
		1,
	)
	require.NoError(t, err)

	cliente.DefinirID(10)

	output := toOutput(&cliente)

	require.Equal(t, uint64(10), output.ID)
	require.Equal(t, "52998224725", output.Documento)
	require.Equal(t, domain.TipoPessoaFisica, output.Tipo)
	require.Equal(t, "João Da Silva", output.Nome)
	require.Equal(t, "joao@email.com", output.Email)
	require.Equal(t, "44999991234", output.Telefone)
	require.True(t, output.Ativo)
	require.True(t, output.RequerAlterarSenha)
}
