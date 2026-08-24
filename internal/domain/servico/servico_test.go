package servico_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

const criadoPorValido uint64 = 1

func TestNewServico_Valido_NasceAtivo(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "Troca de óleo e filtro", 150.50, 60, criadoPorValido)
	require.NoError(t, err)
	require.Equal(t, "Troca de óleo", s.Nome())
	require.Equal(t, "Troca de óleo e filtro", s.Descricao())
	require.Equal(t, 150.50, s.PrecoBase())
	require.Equal(t, 60, s.TempoEstimado().Minutos())
	require.Equal(t, criadoPorValido, s.CriadoPor())
	require.True(t, s.Ativo())
	require.Zero(t, s.ID())
}

func TestNewServico_NomeComEspacos_Normaliza(t *testing.T) {
	s, err := servico.NewServico("  Troca   de  óleo  ", "  revisão  completa ", 100, 30, criadoPorValido)
	require.NoError(t, err)
	require.Equal(t, "Troca de óleo", s.Nome())
	require.Equal(t, "revisão completa", s.Descricao())
}

func TestNewServico_NomeVazio_RetornaErro(t *testing.T) {
	_, err := servico.NewServico("   ", "descrição", 100, 30, criadoPorValido)
	require.ErrorIs(t, err, servico.ErrNomeObrigatorio)
}

func TestNewServico_DescricaoVazia_RetornaErro(t *testing.T) {
	_, err := servico.NewServico("Troca de óleo", "  ", 100, 30, criadoPorValido)
	require.ErrorIs(t, err, servico.ErrDescricaoObrigatoria)
}

func TestNewServico_PrecoNegativo_RetornaErro(t *testing.T) {
	_, err := servico.NewServico("Troca de óleo", "descrição", -0.01, 30, criadoPorValido)
	require.ErrorIs(t, err, servico.ErrPrecoInvalido)
}

func TestNewServico_PrecoZero_EValido(t *testing.T) {
	s, err := servico.NewServico("Diagnóstico", "avaliação gratuita", 0, 30, criadoPorValido)
	require.NoError(t, err)
	require.Equal(t, 0.0, s.PrecoBase())
}

func TestNewServico_TempoInvalido_RetornaErro(t *testing.T) {
	_, err := servico.NewServico("Troca de óleo", "descrição", 100, 0, criadoPorValido)
	require.ErrorIs(t, err, shared.ErrDuracaoEstimadaInvalida)
}

func TestNewServico_CriadoPorZero_RetornaErro(t *testing.T) {
	_, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, 0)
	require.ErrorIs(t, err, servico.ErrCriadoPorObrigatorio)
}

func TestServico_Atualizar_TrocaCamposMutaveis(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição antiga", 100, 30, criadoPorValido)
	require.NoError(t, err)

	err = s.Atualizar("Alinhamento", "alinhamento e balanceamento", 200.75, 90)
	require.NoError(t, err)
	require.Equal(t, "Alinhamento", s.Nome())
	require.Equal(t, "alinhamento e balanceamento", s.Descricao())
	require.Equal(t, 200.75, s.PrecoBase())
	require.Equal(t, 90, s.TempoEstimado().Minutos())
	require.Equal(t, criadoPorValido, s.CriadoPor())
	require.True(t, s.Ativo())
}

func TestServico_Atualizar_DescricaoVazia_RetornaErroENaoAltera(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)

	err = s.Atualizar("Alinhamento", "  ", 200, 90)
	require.ErrorIs(t, err, servico.ErrDescricaoObrigatoria)
	require.Equal(t, "Troca de óleo", s.Nome())
	require.Equal(t, "descrição", s.Descricao())
}

func TestServico_Atualizar_NomeVazio_RetornaErroENaoAltera(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)

	err = s.Atualizar("", "nova descrição", 200, 90)
	require.ErrorIs(t, err, servico.ErrNomeObrigatorio)
	require.Equal(t, "Troca de óleo", s.Nome())
	require.Equal(t, "descrição", s.Descricao())
	require.Equal(t, 100.0, s.PrecoBase())
	require.Equal(t, 30, s.TempoEstimado().Minutos())
}

func TestServico_Atualizar_PrecoNegativo_RetornaErroENaoAltera(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)

	err = s.Atualizar("Alinhamento", "nova descrição", -1, 90)
	require.ErrorIs(t, err, servico.ErrPrecoInvalido)
	require.Equal(t, "Troca de óleo", s.Nome())
	require.Equal(t, 100.0, s.PrecoBase())
}

func TestServico_Atualizar_TempoInvalido_RetornaErroENaoAltera(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)

	err = s.Atualizar("Alinhamento", "nova descrição", 200, 0)
	require.ErrorIs(t, err, shared.ErrDuracaoEstimadaInvalida)
	require.Equal(t, 30, s.TempoEstimado().Minutos())
}

func TestServico_AtribuirID_PreencheID(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)
	require.Zero(t, s.ID())

	s.AtribuirID(7)
	require.Equal(t, uint64(7), s.ID())
}

func TestNewServico_DataCadastroEDataAtualizacao_NascemIguais(t *testing.T) {
	antes := time.Now()
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)
	depois := time.Now()

	require.False(t, s.DataCadastro().Before(antes))
	require.False(t, s.DataCadastro().After(depois))
	require.Equal(t, s.DataCadastro(), s.DataAtualizacao())
}

func TestServico_Atualizar_AtualizaDataAtualizacaoSemMudarDataCadastro(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)
	cadastroOriginal := s.DataCadastro()
	criadoPorOriginal := s.CriadoPor()

	require.NoError(t, s.Atualizar("Alinhamento", "nova descrição", 200, 90))

	require.Equal(t, cadastroOriginal, s.DataCadastro())
	require.False(t, s.DataAtualizacao().Before(cadastroOriginal))
	require.Equal(t, criadoPorOriginal, s.CriadoPor())
}

func TestServico_AtivarInativar(t *testing.T) {
	s, err := servico.NewServico("Troca de óleo", "descrição", 100, 30, criadoPorValido)
	require.NoError(t, err)
	require.True(t, s.Ativo())

	s.Inativar()
	require.False(t, s.Ativo())

	s.Ativar()
	require.True(t, s.Ativo())
}

func TestRestaurarServico_NaoRevalidaEPreservaEstado(t *testing.T) {
	agora := time.Now()
	s := servico.RestaurarServico(
		42,
		"Troca de óleo",
		"descrição",
		150.50,
		shared.RestaurarDuracaoEstimada(60),
		9,
		false,
		agora,
		agora,
	)

	require.Equal(t, uint64(42), s.ID())
	require.Equal(t, "Troca de óleo", s.Nome())
	require.Equal(t, 150.50, s.PrecoBase())
	require.Equal(t, 60, s.TempoEstimado().Minutos())
	require.Equal(t, uint64(9), s.CriadoPor())
	require.False(t, s.Ativo())
}
