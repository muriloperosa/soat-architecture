package orcamento

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestToDomainMapeiaItensDeServicoEPeca(t *testing.T) {
	agora := time.Date(2026, time.October, 12, 12, 0, 0, 0, time.UTC)
	model := OrcamentoModel{
		ID: 100, OrdemServicoID: 42, ValorItemServicos: 200, ValorItemPecas: 150,
		ValorTotal: 350, Observacoes: "obs", CriadoPor: 7, CriadoEm: agora, AtualizadoEm: agora,
		ItensServico: []ItemServicoModel{{ID: 1, OrcamentoID: 100, ServicoID: 5, Quantidade: 2, Valor: 100, TempoEstimadoMinutos: 60}},
		ItensPeca:    []ItemPecaModel{{ID: 2, OrcamentoID: 100, PecaID: 8, Descricao: "Pastilha", Quantidade: 3, Valor: 50}},
	}

	got := toDomain(model)

	require.Equal(t, uint64(100), got.ID())
	require.Equal(t, uint64(42), got.OrdemServicoID())
	require.Len(t, got.ItensServico(), 1)
	require.Len(t, got.ItensPeca(), 1)
	require.Equal(t, uint64(5), got.ItensServico()[0].ServicoID())
	require.Equal(t, 60, got.ItensServico()[0].TempoEstimado().Minutos())
	require.Equal(t, uint64(8), got.ItensPeca()[0].PecaID())
}
