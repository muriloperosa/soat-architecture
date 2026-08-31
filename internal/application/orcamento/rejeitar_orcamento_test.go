package orcamento_test

import (
	"context"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestRejeitarOrcamentoUseCase_RegistraMotivoNoHistorico(t *testing.T) {
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	o := osAguardandoAprovacao(t, 10)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(o, nil)
	repo.EXPECT().Atualizar(mock.Anything, o).Return(nil)

	uc := app.NewRejeitarOrcamentoUseCase(repo)
	out, err := uc.Executar(context.Background(), app.RejeitarOrcamentoInput{
		OrdemServicoID: 42,
		ClienteID:      10,
		Motivo:         "Valor acima do esperado",
	})

	require.NoError(t, err)
	require.Equal(t, "REJEITADA", out.Status)
	historico := o.HistoricoStatus()
	require.Equal(t, "Valor acima do esperado", historico[len(historico)-1].Motivo())
	require.Equal(t, uint64(2), historico[len(historico)-1].AlteradoPor())
}
