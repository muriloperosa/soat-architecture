package ordemservico_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/stretchr/testify/require"
)

func TestStatusOrdemServicoExisteCaminhoValido(t *testing.T) {
	testes := []struct {
		nome     string
		de       ordemservico.StatusOrdemServico
		para     ordemservico.StatusOrdemServico
		esperado bool
	}{
		{"adjacente direto", ordemservico.StatusRecebida, ordemservico.StatusEmDiagnostico, true},
		{"salto de varias etapas", ordemservico.StatusRecebida, ordemservico.StatusEntregue, true},
		{"mesmo status", ordemservico.StatusRecebida, ordemservico.StatusRecebida, false},
		{"sem caminho, retrocesso real", ordemservico.StatusAprovada, ordemservico.StatusRecebida, false},
		{"caminho via ciclo, aguardando para rejeitada", ordemservico.StatusAguardandoAprovacao, ordemservico.StatusRejeitada, true},
		{"caminho via ciclo, rejeitada para aguardando", ordemservico.StatusRejeitada, ordemservico.StatusAguardandoAprovacao, true},
		{"rejeitada alcanca entregue apos reaprovacao", ordemservico.StatusRejeitada, ordemservico.StatusEntregue, true},
		{"status terminal sem saida", ordemservico.StatusEntregue, ordemservico.StatusRecebida, false},
		{"status invalido como origem", ordemservico.StatusOrdemServico("INVALIDO"), ordemservico.StatusEntregue, false},
		{"status invalido como destino", ordemservico.StatusRecebida, ordemservico.StatusOrdemServico("INVALIDO"), false},
	}

	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			require.Equal(t, teste.esperado, teste.de.ExisteCaminhoValido(teste.para))
		})
	}
}
