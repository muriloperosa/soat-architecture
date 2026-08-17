package shared_test

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func TestPapelUsuario_Valido(t *testing.T) {
	require.True(t, shared.PapelAdmin.Valido())
	require.True(t, shared.PapelMecanico.Valido())
	require.True(t, shared.PapelAtendente.Valido())
	require.True(t, shared.PapelCliente.Valido())
	require.False(t, shared.PapelUsuario("gerente").Valido())
}
