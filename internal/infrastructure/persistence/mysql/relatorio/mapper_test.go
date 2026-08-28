package relatorio

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToDomain_ComValores(t *testing.T) {
	resultado := transicaoStatusResultado{
		Total:          3,
		MediaSegundos:  sql.NullFloat64{Valid: true, Float64: 7200},
		MinimaSegundos: sql.NullFloat64{Valid: true, Float64: 3600},
		MaximaSegundos: sql.NullFloat64{Valid: true, Float64: 10800},
	}

	saida := toDomain(resultado)

	require.Equal(t, 3, saida.TotalOrdens)
	require.Equal(t, 2*time.Hour, saida.DuracaoMedia)
	require.Equal(t, time.Hour, saida.DuracaoMinima)
	require.Equal(t, 3*time.Hour, saida.DuracaoMaxima)
}

func TestToDomain_SemLinhas_DuracoesZeradas(t *testing.T) {
	saida := toDomain(transicaoStatusResultado{Total: 0})

	require.Zero(t, saida.TotalOrdens)
	require.Zero(t, saida.DuracaoMedia)
	require.Zero(t, saida.DuracaoMinima)
	require.Zero(t, saida.DuracaoMaxima)
}
