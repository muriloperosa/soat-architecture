package ordemservico

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToDomainRejeitaDadosInvalidos(t *testing.T) {
	base := OrdemServicoModel{Numero: "OS-20261012-a1b2c3d4e5f6", Status: "RECEBIDA"}

	t.Run("número inválido", func(t *testing.T) {
		model := base
		model.Numero = ""
		os, err := toDomain(model)
		require.Nil(t, os)
		require.Error(t, err)
	})

	t.Run("status da ordem inválido", func(t *testing.T) {
		model := base
		model.Status = "DESCONHECIDO"
		os, err := toDomain(model)
		require.Nil(t, os)
		require.Error(t, err)
	})

	t.Run("status do histórico inválido", func(t *testing.T) {
		model := base
		model.Historicos = []HistoricoStatusModel{{Status: "DESCONHECIDO"}}
		os, err := toDomain(model)
		require.Nil(t, os)
		require.Error(t, err)
	})
}
