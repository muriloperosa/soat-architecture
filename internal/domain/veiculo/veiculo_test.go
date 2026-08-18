package veiculo_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	"github.com/stretchr/testify/require"
)

func TestNewVeiculo_Valido_NasceAtivo(t *testing.T) {
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)
	require.Equal(t, "ABC1234", veiculoCriado.Placa().String())
	require.Equal(t, "Fiat", veiculoCriado.Marca())
	require.Equal(t, "Uno", veiculoCriado.Modelo())
	require.Equal(t, uint32(15000), veiculoCriado.QuilometragemAtual())
	require.Equal(t, uint16(2020), veiculoCriado.Ano())
	require.Equal(t, "Branco", veiculoCriado.Cor())
	require.Equal(t, uint64(1), veiculoCriado.CriadoPor())
	require.True(t, veiculoCriado.Ativo())
}

func TestNewVeiculo_PlacaInvalida_RetornaErro(t *testing.T) {
	_, err := veiculo.NewVeiculo("INVALIDA", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.ErrorIs(t, err, shared.ErrPlacaInvalida)
}

func TestNewVeiculo_MarcaVazia_RetornaErro(t *testing.T) {
	_, err := veiculo.NewVeiculo("ABC1234", "", "Uno", 15000, 2020, "Branco", 1)
	require.ErrorIs(t, err, veiculo.ErrMarcaObrigatoria)
}

func TestNewVeiculo_ModeloVazio_RetornaErro(t *testing.T) {
	_, err := veiculo.NewVeiculo("ABC1234", "Fiat", "", 15000, 2020, "Branco", 1)
	require.ErrorIs(t, err, veiculo.ErrModeloObrigatorio)
}

func TestNewVeiculo_CorVazia_RetornaErro(t *testing.T) {
	_, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "", 1)
	require.ErrorIs(t, err, veiculo.ErrCorObrigatoria)
}

func TestNewVeiculo_AnoMenorQueMinimo_RetornaErro(t *testing.T) {
	_, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 1899, "Branco", 1)
	require.ErrorIs(t, err, veiculo.ErrAnoInvalido)
}

func TestNewVeiculo_AnoMuitoFuturo_RetornaErro(t *testing.T) {
	anoFuturo := uint16(time.Now().Year() + 2)
	_, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, anoFuturo, "Branco", 1)
	require.ErrorIs(t, err, veiculo.ErrAnoInvalido)
}

func TestNewVeiculo_CriadoPorZero_RetornaErro(t *testing.T) {
	_, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 0)
	require.ErrorIs(t, err, veiculo.ErrCriadoPorObrigatorio)
}

func TestVeiculo_Atualizar_TrocaMarcaModeloECor(t *testing.T) {
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)

	err = veiculoCriado.Atualizar("Volkswagen", "Gol", "Prata")
	require.NoError(t, err)
	require.Equal(t, "Volkswagen", veiculoCriado.Marca())
	require.Equal(t, "Gol", veiculoCriado.Modelo())
	require.Equal(t, "Prata", veiculoCriado.Cor())
}

func TestVeiculo_Atualizar_MarcaVazia_RetornaErroENaoAltera(t *testing.T) {
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)

	err = veiculoCriado.Atualizar("", "Gol", "Prata")
	require.ErrorIs(t, err, veiculo.ErrMarcaObrigatoria)
	require.Equal(t, "Fiat", veiculoCriado.Marca())
}

func TestVeiculo_AtualizarQuilometragem_Valida(t *testing.T) {
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)

	err = veiculoCriado.AtualizarQuilometragem(16000)
	require.NoError(t, err)
	require.Equal(t, uint32(16000), veiculoCriado.QuilometragemAtual())
}

func TestVeiculo_AtualizarQuilometragem_MenorQueAtual_RetornaErro(t *testing.T) {
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)

	err = veiculoCriado.AtualizarQuilometragem(14000)
	require.ErrorIs(t, err, veiculo.ErrQuilometragemInvalida)
	require.Equal(t, uint32(15000), veiculoCriado.QuilometragemAtual())
}

func TestVeiculo_AtivarInativar(t *testing.T) {
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)

	veiculoCriado.Inativar()
	require.False(t, veiculoCriado.Ativo())

	veiculoCriado.Ativar()
	require.True(t, veiculoCriado.Ativo())
}

func TestVeiculo_AtribuirID_PreencheID(t *testing.T) {
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)
	require.Zero(t, veiculoCriado.ID())

	veiculoCriado.AtribuirID(7)
	require.Equal(t, uint64(7), veiculoCriado.ID())
}

func TestNewVeiculo_DataCadastroEDataAtualizacao_NascemIguais(t *testing.T) {
	antes := time.Now()
	veiculoCriado, err := veiculo.NewVeiculo("ABC1234", "Fiat", "Uno", 15000, 2020, "Branco", 1)
	require.NoError(t, err)
	depois := time.Now()

	require.False(t, veiculoCriado.DataCadastro().Before(antes))
	require.False(t, veiculoCriado.DataCadastro().After(depois))
	require.Equal(t, veiculoCriado.DataCadastro(), veiculoCriado.DataAtualizacao())
}

func TestRestaurarVeiculo_NaoRevalidaEPreservaEstado(t *testing.T) {
	placa, err := shared.NewPlaca("ABC1234")
	require.NoError(t, err)

	agora := time.Now()
	veiculoCriado := veiculo.RestaurarVeiculo(42, placa, "Fiat", "Uno", 15000, 2020, "Branco", 1, false, agora, agora)

	require.Equal(t, uint64(42), veiculoCriado.ID())
	require.Equal(t, "Fiat", veiculoCriado.Marca())
	require.False(t, veiculoCriado.Ativo())
}
