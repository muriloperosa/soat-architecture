package veiculo_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/stretchr/testify/require"
)

func TestNewPlaca_FormatoAntigo_NormalizaEValida(t *testing.T) {
	placaCriada, err := veiculo.NewPlaca("  abc-1234  ")
	require.NoError(t, err)
	require.Equal(t, "ABC1234", placaCriada.String())
}

func TestNewPlaca_FormatoMercosul_NormalizaEValida(t *testing.T) {
	placaCriada, err := veiculo.NewPlaca("abc1d23")
	require.NoError(t, err)
	require.Equal(t, "ABC1D23", placaCriada.String())
}

func TestNewPlaca_Vazia_RetornaErro(t *testing.T) {
	_, err := veiculo.NewPlaca("   ")
	require.ErrorIs(t, err, veiculo.ErrPlacaObrigatoria)
}

func TestNewPlaca_FormatoInvalido_RetornaErro(t *testing.T) {
	_, err := veiculo.NewPlaca("ABCD123")
	require.ErrorIs(t, err, veiculo.ErrPlacaInvalida)
}

func TestNewPlaca_CurtaDemais_RetornaErro(t *testing.T) {
	_, err := veiculo.NewPlaca("ABC123")
	require.ErrorIs(t, err, veiculo.ErrPlacaInvalida)
}
