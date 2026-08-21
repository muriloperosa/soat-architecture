package cliente

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewDocumentoCPFValido(t *testing.T) {
	documento, err := NewDocumento("529.982.247-25", TipoPessoaFisica)

	require.NoError(t, err)
	require.Equal(t, "52998224725", documento.String())
	require.Equal(t, TipoPessoaFisica, documento.Tipo())
	require.Equal(t, "529.982.247-25", documento.Formatado())
	require.False(t, documento.IsZero())
}

func TestNewDocumentoCNPJValido(t *testing.T) {
	documento, err := NewDocumento("04.252.011/0001-10", TipoPessoaJuridica)

	require.NoError(t, err)
	require.Equal(t, "04252011000110", documento.String())
	require.Equal(t, TipoPessoaJuridica, documento.Tipo())
	require.Equal(t, "04.252.011/0001-10", documento.Formatado())
	require.False(t, documento.IsZero())
}

func TestNewDocumentoObrigatorio(t *testing.T) {
	documento, err := NewDocumento("", TipoPessoaFisica)

	require.ErrorIs(t, err, ErrDocumentoObrigatorio)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoObrigatorioAposNormalizacao(t *testing.T) {
	documento, err := NewDocumento("...", TipoPessoaFisica)

	require.ErrorIs(t, err, ErrDocumentoObrigatorio)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoTipoPessoaInvalido(t *testing.T) {
	documento, err := NewDocumento("52998224725", TipoPessoa("X"))

	require.ErrorIs(t, err, ErrTipoPessoaInvalido)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoCPFInvalido(t *testing.T) {
	documento, err := NewDocumento("123.456.789-00", TipoPessoaFisica)

	require.ErrorIs(t, err, ErrCPFInvalido)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoCNPJInvalido(t *testing.T) {
	documento, err := NewDocumento("11.111.111/1111-11", TipoPessoaJuridica)

	require.ErrorIs(t, err, ErrCNPJInvalido)
	require.True(t, documento.IsZero())
}

func TestDocumentoFormatadoComTipoInvalido(t *testing.T) {
	documento := Documento{
		valor: "123456",
		tipo:  TipoPessoa("X"),
	}

	require.Equal(t, "123456", documento.Formatado())
}

func TestValidarCPF(t *testing.T) {
	tests := []struct {
		name     string
		valor    string
		expected bool
	}{
		{
			name:     "deve validar CPF",
			valor:    "52998224725",
			expected: true,
		},
		{
			name:     "deve validar CPF com primeiro digito calculado como zero",
			valor:    "12345678909",
			expected: true,
		},
		{
			name:     "deve validar CPF com segundo digito calculado como zero",
			valor:    "10000000280",
			expected: true,
		},
		{
			name:     "deve rejeitar CPF com tamanho invalido",
			valor:    "5299822472",
			expected: false,
		},
		{
			name:     "deve rejeitar CPF com todos os digitos iguais",
			valor:    "11111111111",
			expected: false,
		},
		{
			name:     "deve rejeitar CPF com primeiro digito verificador invalido",
			valor:    "52998224715",
			expected: false,
		},
		{
			name:     "deve rejeitar CPF com segundo digito verificador invalido",
			valor:    "52998224724",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, validarCPF(tt.valor))
		})
	}
}

func TestValidarCNPJ(t *testing.T) {
	tests := []struct {
		name     string
		valor    string
		expected bool
	}{
		{
			name:     "deve validar CNPJ",
			valor:    "04252011000110",
			expected: true,
		},
		{
			name:     "deve validar CNPJ com segundo resto maior ou igual a dois",
			valor:    "11222333000181",
			expected: true,
		},
		{
			name:     "deve rejeitar CNPJ com tamanho invalido",
			valor:    "0425201100011",
			expected: false,
		},
		{
			name:     "deve rejeitar CNPJ com todos os digitos iguais",
			valor:    "11111111111111",
			expected: false,
		},
		{
			name:     "deve rejeitar CNPJ com primeiro digito verificador invalido",
			valor:    "04252011000120",
			expected: false,
		},
		{
			name:     "deve rejeitar CNPJ com segundo digito verificador invalido",
			valor:    "04252011000111",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, validarCNPJ(tt.valor))
		})
	}
}

func TestTodosDigitosIguais(t *testing.T) {
	tests := []struct {
		name     string
		valor    string
		expected bool
	}{
		{
			name:     "deve retornar true quando todos os digitos forem iguais",
			valor:    "111111",
			expected: true,
		},
		{
			name:     "deve retornar false quando existir digito diferente",
			valor:    "111112",
			expected: false,
		},
		{
			name:     "deve retornar true para um unico digito",
			valor:    "1",
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, todosDigitosIguais(tt.valor))
		})
	}
}

func TestFormatarCPF(t *testing.T) {
	tests := []struct {
		name     string
		valor    string
		expected string
	}{
		{
			name:     "deve formatar CPF",
			valor:    "52998224725",
			expected: "529.982.247-25",
		},
		{
			name:     "deve retornar valor original quando tamanho for invalido",
			valor:    "123",
			expected: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, formatarCPF(tt.valor))
		})
	}
}

func TestFormatarCNPJ(t *testing.T) {
	tests := []struct {
		name     string
		valor    string
		expected string
	}{
		{
			name:     "deve formatar CNPJ",
			valor:    "04252011000110",
			expected: "04.252.011/0001-10",
		},
		{
			name:     "deve retornar valor original quando tamanho for invalido",
			valor:    "123",
			expected: "123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.expected, formatarCNPJ(tt.valor))
		})
	}
}
