package relatorio_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
	"github.com/stretchr/testify/require"
)

func TestNewPeriodo(t *testing.T) {
	agora := time.Now()

	t.Run("caso feliz, dentro dos limites", func(t *testing.T) {
		inicio := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		fim := agora

		periodo, err := relatorio.NewPeriodo(inicio, fim)

		require.NoError(t, err)
		require.True(t, periodo.Inicio().Equal(inicio))
		require.True(t, periodo.Fim().Equal(fim))
	})

	t.Run("inicio igual a fim", func(t *testing.T) {
		momento := time.Date(2026, time.February, 1, 0, 0, 0, 0, time.UTC)

		_, err := relatorio.NewPeriodo(momento, momento)

		require.ErrorIs(t, err, relatorio.ErrPeriodoInicioMaiorOuIgualFim)
	})

	t.Run("inicio depois do fim", func(t *testing.T) {
		inicio := time.Date(2026, time.March, 10, 0, 0, 0, 0, time.UTC)
		fim := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)

		_, err := relatorio.NewPeriodo(inicio, fim)

		require.ErrorIs(t, err, relatorio.ErrPeriodoInicioMaiorOuIgualFim)
	})

	t.Run("fim no futuro", func(t *testing.T) {
		inicio := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		fim := time.Now().Add(24 * time.Hour)

		_, err := relatorio.NewPeriodo(inicio, fim)

		require.ErrorIs(t, err, relatorio.ErrPeriodoFimNoFuturo)
	})

	t.Run("fim igual a agora, limite valido", func(t *testing.T) {
		inicio := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		fim := time.Now()

		_, err := relatorio.NewPeriodo(inicio, fim)

		require.NoError(t, err)
	})

	t.Run("inicio antes do limite minimo", func(t *testing.T) {
		inicio := time.Date(2025, time.December, 31, 23, 59, 59, 0, time.UTC)
		fim := agora

		_, err := relatorio.NewPeriodo(inicio, fim)

		require.ErrorIs(t, err, relatorio.ErrPeriodoInicioAntesDoLimite)
	})

	t.Run("inicio igual ao limite minimo, valido", func(t *testing.T) {
		inicio := time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)
		fim := agora

		_, err := relatorio.NewPeriodo(inicio, fim)

		require.NoError(t, err)
	})
}
