package veiculo_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/stretchr/testify/require"
)

func TestNewCor_Valida_Normaliza(t *testing.T) {
	corCriada, err := veiculo.NewCor("  azul metálico  ")
	require.NoError(t, err)
	require.Equal(t, "Azul Metálico", corCriada.String())
}

func TestNewCor_Vazia_RetornaErro(t *testing.T) {
	_, err := veiculo.NewCor("   ")
	require.ErrorIs(t, err, veiculo.ErrCorObrigatoria)
}

func TestNewCor_FormatoInvalido_RetornaErro(t *testing.T) {
	_, err := veiculo.NewCor("Azul123")
	require.ErrorIs(t, err, veiculo.ErrCorInvalida)
}
