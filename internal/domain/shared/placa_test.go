package shared_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestNewPlaca_FormatoAntigo_NormalizaEValida(t *testing.T) {
	placaCriada, err := shared.NewPlaca("  abc-1234  ")
	require.NoError(t, err)
	require.Equal(t, "ABC1234", placaCriada.String())
}

func TestNewPlaca_FormatoMercosul_NormalizaEValida(t *testing.T) {
	placaCriada, err := shared.NewPlaca("abc1d23")
	require.NoError(t, err)
	require.Equal(t, "ABC1D23", placaCriada.String())
}

func TestNewPlaca_Vazia_RetornaErro(t *testing.T) {
	_, err := shared.NewPlaca("   ")
	require.ErrorIs(t, err, shared.ErrPlacaObrigatoria)
}

func TestNewPlaca_FormatoInvalido_RetornaErro(t *testing.T) {
	_, err := shared.NewPlaca("ABCD123")
	require.ErrorIs(t, err, shared.ErrPlacaInvalida)
}

func TestNewPlaca_CurtaDemais_RetornaErro(t *testing.T) {
	_, err := shared.NewPlaca("ABC123")
	require.ErrorIs(t, err, shared.ErrPlacaInvalida)
}
