package cliente

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewTelefoneCelularValido(t *testing.T) {
	telefone, err := NewTelefone("(44) 99999-1234")
	require.NoError(t, err)
	require.Equal(t, "44999991234", telefone.String())
	require.Equal(t, "(44) 99999-1234", telefone.Formatado())
}

func TestNewTelefoneCelularSemMascara(t *testing.T) {
	telefone, err := NewTelefone("44999991234")
	require.NoError(t, err)
	require.Equal(t, "44999991234", telefone.String())
}

func TestNewTelefoneFixoValido(t *testing.T) {
	telefone, err := NewTelefone("(44) 3031-1234")

	require.NoError(t, err)
	require.Equal(t, "4430311234", telefone.String())
	require.Equal(t, "(44) 3031-1234", telefone.Formatado())
}

func TestNewTelefoneObrigatorio(t *testing.T) {
	_, err := NewTelefone("")

	require.Error(t, err)
	require.ErrorIs(t, err, ErrTelefoneObrigatorio)
}

func TestNewTelefoneApenasCaracteresInvalidos(t *testing.T) {
	_, err := NewTelefone("abc() -")

	require.ErrorIs(t, err, ErrTelefoneObrigatorio)
}

func TestNewTelefoneComMenosDeDezDigitos(t *testing.T) {
	_, err := NewTelefone("449999123")

	require.ErrorIs(t, err, ErrTelefoneInvalido)
}

func TestNewTelefoneComMaisDeOnzeDigitos(t *testing.T) {
	_, err := NewTelefone("449999912345")

	require.ErrorIs(t, err, ErrTelefoneInvalido)
}

func TestNewTelefoneComDDDIniciadoEmZero(t *testing.T) {
	_, err := NewTelefone("(04) 99999-1234")

	require.ErrorIs(t, err, ErrTelefoneInvalido)
}

func TestNewTelefoneCelularSemNonoDigito(t *testing.T) {
	_, err := NewTelefone("(44) 89999-1234")

	require.ErrorIs(t, err, ErrTelefoneInvalido)
}

func TestNewTelefoneFixoComPrefixoInvalido(t *testing.T) {
	_, err := NewTelefone("(44) 9031-1234")

	require.ErrorIs(t, err, ErrTelefoneInvalido)
}

func TestTelefoneFormatado_DeveRetornarValorOriginalQuandoTamanhoForInvalido(t *testing.T) {
	telefone := Telefone{valor: "123"}
	result := telefone.Formatado()
	require.Equal(t, "123", result)
}
