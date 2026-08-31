package orcamento_test

import (
	"strings"
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestNewOrcamento(t *testing.T) {
	antes := time.Now()
	o, err := orcamento.NewOrcamento(10, "  orçamento inicial  ", 30)
	depois := time.Now()

	require.NoError(t, err)
	require.Zero(t, o.ID())
	require.Equal(t, uint64(10), o.OrdemServicoID())
	require.Equal(t, "orçamento inicial", o.Observacoes())
	require.Equal(t, uint64(30), o.CriadoPor())
	require.Zero(t, o.ValorTotal())
	require.Empty(t, o.ItensServico())
	require.Empty(t, o.ItensPeca())
	require.False(t, o.CriadoEm().Before(antes))
	require.False(t, o.CriadoEm().After(depois))
	require.Equal(t, o.CriadoEm(), o.AtualizadoEm())
}

func TestNewOrcamentoValidaInvariantes(t *testing.T) {
	testes := []struct {
		nome           string
		ordemServicoID uint64
		criadoPor      uint64
		erroEsperado   error
	}{
		{"ordem de serviço obrigatória", 0, 3, orcamento.ErrOrdemServicoObrigatoria},
		{"criado por obrigatório", 10, 0, orcamento.ErrCriadoPorObrigatorio},
	}

	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			o, err := orcamento.NewOrcamento(teste.ordemServicoID, "", teste.criadoPor)

			require.Nil(t, o)
			require.ErrorIs(t, err, teste.erroEsperado)
		})
	}
}

func TestNewOrcamento_ObservacoesExcedeTamanhoMaximo(t *testing.T) {
	o, err := orcamento.NewOrcamento(10, strings.Repeat("a", 501), 3)

	require.Nil(t, o)
	require.ErrorIs(t, err, orcamento.ErrObservacoesInvalidas)
}

func TestNewOrcamento_ObservacoesNoLimite(t *testing.T) {
	o, err := orcamento.NewOrcamento(10, strings.Repeat("a", 500), 3)

	require.NoError(t, err)
	require.Len(t, o.Observacoes(), 500)
}

func TestAdicionarItemServico_CalculaTotais(t *testing.T) {
	o, err := orcamento.NewOrcamento(10, "", 3)
	require.NoError(t, err)

	err = o.AdicionarItemServico(1, 2, 100.0, 60)
	require.NoError(t, err)

	require.Len(t, o.ItensServico(), 1)
	require.Equal(t, 200.0, o.ValorItemServicos())
	require.Zero(t, o.ValorItemPecas())
	require.Equal(t, 200.0, o.ValorTotal())

	item := o.ItensServico()[0]
	require.Equal(t, uint64(1), item.ServicoID())
	require.Equal(t, 2, item.Quantidade())
	require.Equal(t, 100.0, item.Valor())
	require.Equal(t, 60, item.TempoEstimado().Minutos())
	require.Equal(t, 200.0, item.CalcularSubtotal())
}

func TestAdicionarItemServico_ValidaDados(t *testing.T) {
	o, err := orcamento.NewOrcamento(10, "", 3)
	require.NoError(t, err)

	require.ErrorIs(t, o.AdicionarItemServico(0, 1, 10, 10), orcamento.ErrServicoObrigatorio)
	require.ErrorIs(t, o.AdicionarItemServico(1, 0, 10, 10), orcamento.ErrQuantidadeInvalida)
	require.ErrorIs(t, o.AdicionarItemServico(1, 1, -10, 10), orcamento.ErrValorInvalido)
	require.ErrorIs(t, o.AdicionarItemServico(1, 1, 10, 0), shared.ErrDuracaoEstimadaInvalida)
	require.Empty(t, o.ItensServico())
}

func TestAdicionarItemPeca_CalculaTotais(t *testing.T) {
	o, err := orcamento.NewOrcamento(10, "", 3)
	require.NoError(t, err)

	err = o.AdicionarItemPeca(5, "  Filtro de óleo  ", 3, 50.0)
	require.NoError(t, err)

	require.Len(t, o.ItensPeca(), 1)
	require.Zero(t, o.ValorItemServicos())
	require.Equal(t, 150.0, o.ValorItemPecas())
	require.Equal(t, 150.0, o.ValorTotal())

	item := o.ItensPeca()[0]
	require.Equal(t, uint64(5), item.PecaID())
	require.Equal(t, "Filtro de óleo", item.Descricao())
	require.Equal(t, 3, item.Quantidade())
	require.Equal(t, 50.0, item.Valor())
	require.Equal(t, 150.0, item.CalcularSubtotal())
}

func TestAdicionarItemPeca_ValidaDados(t *testing.T) {
	o, err := orcamento.NewOrcamento(10, "", 3)
	require.NoError(t, err)

	require.ErrorIs(t, o.AdicionarItemPeca(0, "Filtro", 1, 10), orcamento.ErrPecaObrigatoria)
	require.ErrorIs(t, o.AdicionarItemPeca(1, "   ", 1, 10), orcamento.ErrDescricaoObrigatoria)
	require.ErrorIs(t, o.AdicionarItemPeca(1, "Filtro", 0, 10), orcamento.ErrQuantidadeInvalida)
	require.ErrorIs(t, o.AdicionarItemPeca(1, "Filtro", 1, -10), orcamento.ErrValorInvalido)
	require.ErrorIs(t, o.AdicionarItemPeca(1, strings.Repeat("a", 501), 1, 10), orcamento.ErrDescricaoInvalida)
	require.Empty(t, o.ItensPeca())
}

func TestOrcamentoTotais_ServicosMaisPecas(t *testing.T) {
	o, err := orcamento.NewOrcamento(10, "", 3)
	require.NoError(t, err)

	require.NoError(t, o.AdicionarItemServico(1, 2, 100.0, 60))
	require.NoError(t, o.AdicionarItemPeca(5, "Filtro", 3, 50.0))

	require.Equal(t, 200.0, o.ValorItemServicos())
	require.Equal(t, 150.0, o.ValorItemPecas())
	require.Equal(t, 350.0, o.ValorTotal())
	require.Equal(t, 350.0, o.CalcularTotal())
}

func orcamentoComItens(t *testing.T) *orcamento.Orcamento {
	t.Helper()

	itensServico := []orcamento.ItemServico{
		orcamento.ReidratarItemServico(1, 42, 1, 2, 100.0, shared.RestaurarDuracaoEstimada(60)),
	}
	itensPeca := []orcamento.ItemPeca{
		orcamento.ReidratarItemPeca(1, 42, 5, "Filtro", 3, 50.0),
	}

	cadastro := time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC)
	return orcamento.ReidratarOrcamento(
		42, 10,
		itensServico, itensPeca,
		200.0, 150.0, 350.0,
		"observação",
		3,
		cadastro, cadastro,
	)
}

func TestReidratarOrcamento(t *testing.T) {
	o := orcamentoComItens(t)

	require.Equal(t, uint64(42), o.ID())
	require.Equal(t, uint64(10), o.OrdemServicoID())
	require.Equal(t, 200.0, o.ValorItemServicos())
	require.Equal(t, 150.0, o.ValorItemPecas())
	require.Equal(t, 350.0, o.ValorTotal())
	require.Equal(t, "observação", o.Observacoes())
	require.Len(t, o.ItensServico(), 1)
	require.Len(t, o.ItensPeca(), 1)
}

func TestRemoverItemServico_ComSucesso(t *testing.T) {
	o := orcamentoComItens(t)

	err := o.RemoverItemServico(1)

	require.NoError(t, err)
	require.Empty(t, o.ItensServico())
	require.Zero(t, o.ValorItemServicos())
	require.Equal(t, 150.0, o.ValorTotal())
}

func TestRemoverItemServico_NaoEncontrado(t *testing.T) {
	o := orcamentoComItens(t)

	err := o.RemoverItemServico(999)

	require.ErrorIs(t, err, orcamento.ErrItemServicoNaoEncontrado)
	require.Len(t, o.ItensServico(), 1)
}

func TestRemoverItemPeca_ComSucesso(t *testing.T) {
	o := orcamentoComItens(t)

	err := o.RemoverItemPeca(1)

	require.NoError(t, err)
	require.Empty(t, o.ItensPeca())
	require.Zero(t, o.ValorItemPecas())
	require.Equal(t, 200.0, o.ValorTotal())
}

func TestRemoverItemPeca_NaoEncontrado(t *testing.T) {
	o := orcamentoComItens(t)

	err := o.RemoverItemPeca(999)

	require.ErrorIs(t, err, orcamento.ErrItemPecaNaoEncontrado)
	require.Len(t, o.ItensPeca(), 1)
}

func TestOrcamento_ValidarParaEnvio_OrcamentoVazioRetornaErro(t *testing.T) {
	o, err := orcamento.NewOrcamento(1, "", 2)
	require.NoError(t, err)

	require.ErrorIs(t, o.ValidarParaEnvio(), orcamento.ErrOrcamentoVazio)
}

func TestOrcamento_ValidarParaEnvio_ComItemValido(t *testing.T) {
	o, err := orcamento.NewOrcamento(1, "", 2)
	require.NoError(t, err)
	require.NoError(t, o.AdicionarItemServico(3, 1, 100, 60))

	require.NoError(t, o.ValidarParaEnvio())
	require.Equal(t, 100.0, o.ValorTotal())
}

func TestOrcamento_AlterarQuantidadeItemPeca_RecalculaTotais(t *testing.T) {
	agora := time.Now()
	o := orcamento.ReidratarOrcamento(
		1, 1,
		nil,
		[]orcamento.ItemPeca{orcamento.ReidratarItemPeca(11, 1, 7, "Pastilha", 2, 50)},
		0, 100, 100, "", 10, agora, agora,
	)

	require.NoError(t, o.AlterarQuantidadeItemPeca(11, 4))
	require.Equal(t, 4, o.ItensPeca()[0].Quantidade())
	require.Equal(t, 200.0, o.ValorItemPecas())
	require.Equal(t, 200.0, o.ValorTotal())
}

func TestOrcamento_AlterarQuantidadeItemPeca_QuantidadeInvalida(t *testing.T) {
	o := orcamento.ReidratarOrcamento(
		1, 1,
		nil,
		[]orcamento.ItemPeca{orcamento.ReidratarItemPeca(11, 1, 7, "Pastilha", 2, 50)},
		0, 100, 100, "", 10, time.Now(), time.Now(),
	)

	err := o.AlterarQuantidadeItemPeca(11, 0)
	require.ErrorIs(t, err, orcamento.ErrQuantidadeInvalida)
	require.Equal(t, 2, o.ItensPeca()[0].Quantidade())
}
