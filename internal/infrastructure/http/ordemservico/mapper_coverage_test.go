package ordemservico

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestToOrcamentoResumoResponseNil(t *testing.T) {
	require.Nil(t, toOrcamentoResumoResponse(nil))
}
