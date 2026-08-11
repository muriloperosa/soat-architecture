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
}

func TestNewDocumentoCPFInvalido(t *testing.T) {
	documento, err := NewDocumento("123.456.789-00", TipoPessoaFisica)
	require.ErrorIs(t, err, ErrCPFInvalido)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoCNPJValido(t *testing.T) {
	documento, err := NewDocumento("04.252.011/0001-10", TipoPessoaJuridica)
	require.NoError(t, err)
	require.Equal(t, "04252011000110", documento.String())
	require.Equal(t, TipoPessoaJuridica, documento.Tipo())
	require.Equal(t, "04.252.011/0001-10", documento.Formatado())
}

func TestNewDocumentoCNPJInvalido(t *testing.T) {
	documento, err := NewDocumento("11.111.111/1111-11",TipoPessoaJuridica)
	require.ErrorIs(t, err, ErrCNPJInvalido)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoObrigatorio(t *testing.T) {
	documento, err := NewDocumento("",TipoPessoaFisica)
	require.ErrorIs(t, err, ErrDocumentoObrigatorio)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoTipoPessoaInvalido(t *testing.T) {
	documento, err := NewDocumento("52998224725",TipoPessoa("X"))
	require.ErrorIs(t, err, ErrTipoPessoaInvalido)
	require.True(t, documento.IsZero())
}

func TestNewDocumentoRejeitaCPFTodosDigitosIguais(t *testing.T) {
	documento, err := NewDocumento("111.111.111-11",TipoPessoaFisica)
	require.ErrorIs(t, err, ErrCPFInvalido)
	require.True(t, documento.IsZero())
}
