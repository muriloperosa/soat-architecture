package peca_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/stretchr/testify/require"
)

func TestNewPeca_Valido_NasceAtivaComCodigo(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)

	require.NoError(t, err)
	require.Equal(t, "Peca 1", p.Nome())
	require.Equal(t, "Marca 1", p.Marca())
	require.Equal(t, "Descricao 1", p.Descricao())
	require.Equal(t, 100.0, p.Preco())
	require.Equal(t, 10, p.QuantidadeEmEstoque())
	require.Equal(t, 5, p.EstoqueMinimo())
	require.Equal(t, uint64(1), p.CriadoPor())
	require.True(t, p.Ativo())
	require.NotEmpty(t, p.Codigo())
	require.Zero(t, p.ID())
}

func TestNewPeca_NomeVazio_RetornaErro(t *testing.T) {
	_, err := peca.NewPeca("", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.ErrorIs(t, err, peca.ErrNomeObrigatorio)
}

func TestNewPeca_MarcaVazia_RetornaErro(t *testing.T) {
	_, err := peca.NewPeca("Peca 1", "", "Descricao 1", 100.0, 10, 5, 1)
	require.ErrorIs(t, err, peca.ErrMarcaObrigatoria)
}

func TestNewPeca_DescricaoVazia_RetornaErro(t *testing.T) {
	_, err := peca.NewPeca("Peca 1", "Marca 1", "", 100.0, 10, 5, 1)
	require.ErrorIs(t, err, peca.ErrDescricaoObrigatoria)
}

func TestNewPeca_PrecoNegativo_RetornaErro(t *testing.T) {
	_, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", -1, 10, 5, 1)
	require.ErrorIs(t, err, peca.ErrPrecoInvalido)
}

func TestNewPeca_QuantidadeEmEstoqueNegativa_RetornaErro(t *testing.T) {
	_, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, -1, 5, 1)
	require.ErrorIs(t, err, peca.ErrQuantidadeEmEstoqueInvalida)
}

func TestNewPeca_EstoqueMinimoNegativo_RetornaErro(t *testing.T) {
	_, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, -1, 1)
	require.ErrorIs(t, err, peca.ErrEstoqueMinimoInvalido)
}

func TestNewPeca_EstoqueMinimoMaiorQueQuantidade_Permitido(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 0, 5, 1)
	require.NoError(t, err)
	require.Equal(t, 0, p.QuantidadeEmEstoque())
	require.Equal(t, 5, p.EstoqueMinimo())
}

func TestNewPeca_CriadoPorZero_RetornaErro(t *testing.T) {
	_, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 0)
	require.ErrorIs(t, err, peca.ErrCriadoPorObrigatorio)
}

func TestNewPeca_DataCadastroEDataAtualizacao_NascemIguais(t *testing.T) {
	antes := time.Now()
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)
	depois := time.Now()

	require.False(t, p.DataCadastro().Before(antes))
	require.False(t, p.DataCadastro().After(depois))
	require.Equal(t, p.DataCadastro(), p.DataAtualizacao())
}

func TestPeca_Atualizar_TrocaDadosCadastraisEstoqueMinimo(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Atualizar("Peca 2", "Marca 2", "Descricao 2", 200.0, 8)
	require.NoError(t, err)
	require.Equal(t, "Peca 2", p.Nome())
	require.Equal(t, "Marca 2", p.Marca())
	require.Equal(t, "Descricao 2", p.Descricao())
	require.Equal(t, 200.0, p.Preco())
	require.Equal(t, 8, p.EstoqueMinimo())
	require.Equal(t, 10, p.QuantidadeEmEstoque())
}

func TestPeca_Atualizar_NomeVazio_RetornaErroENaoAltera(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Atualizar("", "Marca 2", "Descricao 2", 200.0, 8)
	require.ErrorIs(t, err, peca.ErrNomeObrigatorio)
	require.Equal(t, "Peca 1", p.Nome())
}

func TestPeca_Atualizar_PrecoNegativo_RetornaErro(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Atualizar("Peca 1", "Marca 1", "Descricao 1", -1, 8)
	require.ErrorIs(t, err, peca.ErrPrecoInvalido)
}

func TestPeca_Atualizar_EstoqueMinimoNegativo_RetornaErro(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Atualizar("Peca 1", "Marca 1", "Descricao 1", 100.0, -1)
	require.ErrorIs(t, err, peca.ErrEstoqueMinimoInvalido)
}

func TestPeca_Consumir_BaixaEstoque(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Consumir(3)
	require.NoError(t, err)
	require.Equal(t, 7, p.QuantidadeEmEstoque())
}

func TestPeca_Consumir_QuantidadeInvalida_RetornaErro(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Consumir(0)
	require.ErrorIs(t, err, peca.ErrQuantidadeInvalida)
}

func TestPeca_Consumir_DeixariaAbaixoDoMinimo_RetornaErro(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Consumir(6)
	require.ErrorIs(t, err, peca.ErrEstoqueInsuficiente)
	require.Equal(t, 10, p.QuantidadeEmEstoque())
}

func TestPeca_Consumir_ExatamenteAteOMinimo_Permitido(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Consumir(5)
	require.NoError(t, err)
	require.Equal(t, 5, p.QuantidadeEmEstoque())
}

func TestPeca_Repor_AumentaEstoque(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Repor(5)
	require.NoError(t, err)
	require.Equal(t, 15, p.QuantidadeEmEstoque())
}

func TestPeca_Repor_QuantidadeInvalida_RetornaErro(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	err = p.Repor(-1)
	require.ErrorIs(t, err, peca.ErrQuantidadeInvalida)
}

func TestPeca_AtivarInativar(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)

	p.Inativar()
	require.False(t, p.Ativo())

	p.Ativar()
	require.True(t, p.Ativo())
}

func TestPeca_AtribuirID_PreencheID(t *testing.T) {
	p, err := peca.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1)
	require.NoError(t, err)
	require.Zero(t, p.ID())

	p.AtribuirID(7)
	require.Equal(t, uint64(7), p.ID())
}

func TestRestaurarPeca_NaoRevalidaEPreservaEstado(t *testing.T) {
	agora := time.Now()
	p := peca.RestaurarPeca(42, "P123", "Peca 1", "Marca 1", "Descricao 1", 100.0, 10, 5, 1, false, agora, agora)

	require.Equal(t, uint64(42), p.ID())
	require.Equal(t, "P123", p.Codigo())
	require.Equal(t, "Peca 1", p.Nome())
	require.False(t, p.Ativo())
}
