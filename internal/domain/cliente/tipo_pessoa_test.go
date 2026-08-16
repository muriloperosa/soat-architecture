package cliente

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTipoPessoaIsValid(t *testing.T) {
	tests := []struct {
		name     string
		tipo     TipoPessoa
		expected bool
	}{
		{
			name:     "deve retornar true para pessoa fisica",
			tipo:     TipoPessoaFisica,
			expected: true,
		},
		{
			name:     "deve retornar true para pessoa juridica",
			tipo:     TipoPessoaJuridica,
			expected: true,
		},
		{
			name:     "deve retornar false para tipo invalido",
			tipo:     TipoPessoa("XX"),
			expected: false,
		},
		{
			name:     "deve retornar false para tipo vazio",
			tipo:     TipoPessoa(""),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.tipo.IsValid()
			require.Equal(t, tt.expected, result)
		})
	}
}
