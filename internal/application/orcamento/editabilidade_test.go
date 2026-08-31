package orcamento_test

import (
	"context"
	"testing"

	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/stretchr/testify/mock"
)

// ordemServicoRepoEditavel mantém os testes unitários focados no caso de uso
// específico, fornecendo uma OS EM_DIAGNOSTICO para a política de edição.
func ordemServicoRepoEditavel(t *testing.T) *ordemservicomocks.OrdemServicoRepository {
	t.Helper()
	repo := ordemservicomocks.NewOrdemServicoRepository(t)
	repo.EXPECT().BuscarPorID(mock.Anything, mock.Anything).RunAndReturn(
		func(_ context.Context, id uint64) (*domainordemservico.OrdemServico, error) {
			o, err := domainordemservico.NewOrdemServico("OS-TESTE", 10, 20, 0, "", "", 1)
			if err != nil {
				return nil, err
			}
			o.AtribuirID(id)
			if err := o.IniciarDiagnostico(1); err != nil {
				return nil, err
			}
			return o, nil
		},
	)
	return repo
}
