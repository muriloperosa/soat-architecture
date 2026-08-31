package reservapeca

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewReservaPeca_ComSucesso(t *testing.T) {
	r, err := NewReservaPeca(10, 20, 3)

	require.NoError(t, err)
	require.Equal(t, uint64(10), r.OrdemServicoID())
	require.Equal(t, uint64(20), r.PecaID())
	require.Equal(t, 3, r.Quantidade())
	require.False(t, r.CriadaEm().IsZero())
	require.False(t, r.AtualizadaEm().IsZero())
}

func TestNewReservaPeca_QuantidadeInvalida(t *testing.T) {
	_, err := NewReservaPeca(10, 20, 0)
	require.ErrorIs(t, err, ErrQuantidadeInvalida)
}

func TestNewReservaPeca_OrdemServicoObrigatoria(t *testing.T) {
	_, err := NewReservaPeca(0, 20, 1)
	require.ErrorIs(t, err, ErrOrdemServicoObrigatoria)
}

func TestNewReservaPeca_PecaObrigatoria(t *testing.T) {
	_, err := NewReservaPeca(10, 0, 1)
	require.ErrorIs(t, err, ErrPecaObrigatoria)
}

func TestReservaPeca_AtribuirID(t *testing.T) {
	r, err := NewReservaPeca(10, 20, 3)
	require.NoError(t, err)

	r.AtribuirID(99)

	require.Equal(t, uint64(99), r.ID())
}

func TestRestaurarReservaPeca_ReidrataDadosPersistidos(t *testing.T) {
	criadaEm := time.Now().Add(-time.Hour)
	atualizadaEm := time.Now()

	r := RestaurarReservaPeca(7, 10, 20, 4, criadaEm, atualizadaEm)

	require.Equal(t, uint64(7), r.ID())
	require.Equal(t, uint64(10), r.OrdemServicoID())
	require.Equal(t, uint64(20), r.PecaID())
	require.Equal(t, 4, r.Quantidade())
	require.Equal(t, criadaEm, r.CriadaEm())
	require.Equal(t, atualizadaEm, r.AtualizadaEm())
}
