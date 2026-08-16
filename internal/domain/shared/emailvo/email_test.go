package emailvo

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewEmailValido(t *testing.T) {
	email, err := NewEmail("  TESTE@EMAIL.COM  ")

	require.NoError(t, err)
	require.Equal(t, "teste@email.com", email.String())
}

func TestNewEmailObrigatorio(t *testing.T) {
	email, err := NewEmail("   ")

	require.ErrorIs(t, err, ErrEmailObrigatorio)
	require.Equal(t, "", email.String())
}

func TestNewEmailInvalido(t *testing.T) {
	email, err := NewEmail("email-invalido")

	require.ErrorIs(t, err, ErrEmailInvalido)
	require.Equal(t, "", email.String())
}

func TestValidateEmailValido(t *testing.T) {
	err := validateEmail("teste@email.com")

	require.NoError(t, err)
}

func TestValidateEmailDeveRetornarErroQuandoParseAddressFalhar(t *testing.T) {
	err := validateEmail("email-invalido")

	require.ErrorIs(t, err, ErrEmailInvalido)
}

func TestValidateEmailDeveRetornarErroQuandoAddressForDiferenteDoValor(t *testing.T) {
	err := validateEmail("Teste <teste@email.com>")

	require.ErrorIs(t, err, ErrEmailInvalido)
}