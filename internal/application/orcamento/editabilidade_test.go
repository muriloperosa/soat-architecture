package orcamento_test

import (
	"context"

	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
)

// ordemServicoRepoEditavel mantém os testes unitários focados no caso de uso
// específico, fornecendo uma OS EM_DIAGNOSTICO para a política de edição.
type ordemServicoRepoEditavel struct{}

func (ordemServicoRepoEditavel) Salvar(context.Context, *domainordemservico.OrdemServico) error {
	return nil
}
func (ordemServicoRepoEditavel) Atualizar(context.Context, *domainordemservico.OrdemServico) error {
	return nil
}
func (ordemServicoRepoEditavel) BuscarPorNumero(context.Context, string) (*domainordemservico.OrdemServico, error) {
	return nil, domainordemservico.ErrOrdemServicoNaoEncontrada
}
func (ordemServicoRepoEditavel) Listar(context.Context, domainquery.Params) (domainquery.Page[*domainordemservico.OrdemServico], error) {
	return domainquery.Page[*domainordemservico.OrdemServico]{}, nil
}
func (ordemServicoRepoEditavel) BuscarPorID(_ context.Context, id uint64) (*domainordemservico.OrdemServico, error) {
	o, err := domainordemservico.NewOrdemServico("OS-TESTE", 10, 20, 0, "", "", 1)
	if err != nil {
		return nil, err
	}
	o.AtribuirID(id)
	if err := o.IniciarDiagnostico(1); err != nil {
		return nil, err
	}
	return o, nil
}
