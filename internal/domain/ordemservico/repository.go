package ordemservico

import "context"

// OrdemServicoRepository persiste e consulta Ordens de Serviço.
type OrdemServicoRepository interface {
	Salvar(ctx context.Context, ordemServico *OrdemServico) error
	BuscarPorID(ctx context.Context, id uint64) (*OrdemServico, error)
	BuscarPorNumero(ctx context.Context, numero string) (*OrdemServico, error)
	Atualizar(ctx context.Context, ordemServico *OrdemServico) error
}
