package shared_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestNewEmail_Valido_NormalizaEValida(t *testing.T) {
	e, err := shared.NewEmail("  Ana@Oficina.COM  ")
	require.NoError(t, err)
	require.Equal(t, "ana@oficina.com", e.String())
}

func TestNewEmail_Vazio_RetornaErro(t *testing.T) {
	_, err := shared.NewEmail("   ")
	require.ErrorIs(t, err, shared.ErrEmailObrigatorio)
}

func TestNewEmail_FormatoInvalido_RetornaErro(t *testing.T) {
	_, err := shared.NewEmail("nao-e-email")
	require.ErrorIs(t, err, shared.ErrEmailInvalido)
}
