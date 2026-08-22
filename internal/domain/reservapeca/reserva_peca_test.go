package reservapeca_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	"github.com/stretchr/testify/require"
)

func TestNewReservaPeca_Valida_Nasce(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 3)

	require.NoError(t, err)
	require.Equal(t, uint64(1), r.OrdemServicoID())
	require.Equal(t, uint64(2), r.PecaID())
	require.Equal(t, 3, r.Quantidade())
	require.Zero(t, r.ID())
}

func TestNewReservaPeca_OrdemServicoZero_RetornaErro(t *testing.T) {
	_, err := reservapeca.NewReservaPeca(0, 2, 3)
	require.ErrorIs(t, err, reservapeca.ErrOrdemServicoObrigatoria)
}

func TestNewReservaPeca_PecaZero_RetornaErro(t *testing.T) {
	_, err := reservapeca.NewReservaPeca(1, 0, 3)
	require.ErrorIs(t, err, reservapeca.ErrPecaObrigatoria)
}

func TestNewReservaPeca_QuantidadeZeroOuNegativa_RetornaErro(t *testing.T) {
	_, err := reservapeca.NewReservaPeca(1, 2, 0)
	require.ErrorIs(t, err, reservapeca.ErrQuantidadeInvalida)

	_, err = reservapeca.NewReservaPeca(1, 2, -1)
	require.ErrorIs(t, err, reservapeca.ErrQuantidadeInvalida)
}

func TestNewReservaPeca_DataCriacaoEAtualizacao_NascemIguais(t *testing.T) {
	antes := time.Now()
	r, err := reservapeca.NewReservaPeca(1, 2, 3)
	require.NoError(t, err)
	depois := time.Now()

	require.False(t, r.CriadaEm().Before(antes))
	require.False(t, r.CriadaEm().After(depois))
	require.Equal(t, r.CriadaEm(), r.AtualizadaEm())
}

func TestReservaPeca_Aumentar_SomaQuantidade(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 2)
	require.NoError(t, err)

	err = r.Aumentar(3)
	require.NoError(t, err)
	require.Equal(t, 5, r.Quantidade())
}

func TestReservaPeca_Aumentar_QuantidadeInvalida_RetornaErroENaoAltera(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 3)
	require.NoError(t, err)

	err = r.Aumentar(0)
	require.ErrorIs(t, err, reservapeca.ErrQuantidadeInvalida)
	require.Equal(t, 3, r.Quantidade())
}

func TestReservaPeca_Reduzir_SubtraiQuantidade(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 5)
	require.NoError(t, err)

	err = r.Reduzir(2)
	require.NoError(t, err)
	require.Equal(t, 3, r.Quantidade())
}

func TestReservaPeca_Reduzir_ExatamenteAteZero_Permitido(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 5)
	require.NoError(t, err)

	err = r.Reduzir(5)
	require.NoError(t, err)
	require.Zero(t, r.Quantidade())
}

func TestReservaPeca_Reduzir_QuantidadeInvalida_RetornaErroENaoAltera(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 3)
	require.NoError(t, err)

	err = r.Reduzir(0)
	require.ErrorIs(t, err, reservapeca.ErrQuantidadeInvalida)
	require.Equal(t, 3, r.Quantidade())
}

func TestReservaPeca_Reduzir_MaiorQueReservada_RetornaErroENaoAltera(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 3)
	require.NoError(t, err)

	err = r.Reduzir(4)
	require.ErrorIs(t, err, reservapeca.ErrQuantidadeSuperiorAReservada)
	require.Equal(t, 3, r.Quantidade())
}

func TestReservaPeca_AtribuirID_PreencheID(t *testing.T) {
	r, err := reservapeca.NewReservaPeca(1, 2, 3)
	require.NoError(t, err)
	require.Zero(t, r.ID())

	r.AtribuirID(7)
	require.Equal(t, uint64(7), r.ID())
}

func TestRestaurarReservaPeca_NaoRevalidaEPreservaEstado(t *testing.T) {
	agora := time.Now()
	r := reservapeca.RestaurarReservaPeca(42, 1, 2, 3, agora, agora)

	require.Equal(t, uint64(42), r.ID())
	require.Equal(t, uint64(1), r.OrdemServicoID())
	require.Equal(t, uint64(2), r.PecaID())
	require.Equal(t, 3, r.Quantidade())
}
