package shared_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestNewDuracaoEstimada_Valida_GuardaMinutos(t *testing.T) {
	d, err := shared.NewDuracaoEstimada(90)
	require.NoError(t, err)
	require.Equal(t, 90, d.Minutos())
	require.Equal(t, 1.5, d.Horas())
}

func TestNewDuracaoEstimada_Zero_RetornaErro(t *testing.T) {
	_, err := shared.NewDuracaoEstimada(0)
	require.ErrorIs(t, err, shared.ErrDuracaoEstimadaInvalida)
}

func TestNewDuracaoEstimada_Negativa_RetornaErro(t *testing.T) {
	_, err := shared.NewDuracaoEstimada(-1)
	require.ErrorIs(t, err, shared.ErrDuracaoEstimadaInvalida)
}

func TestRestaurarDuracaoEstimada_NaoRevalida(t *testing.T) {
	d := shared.RestaurarDuracaoEstimada(45)
	require.Equal(t, 45, d.Minutos())
}
