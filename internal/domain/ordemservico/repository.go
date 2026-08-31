package ordemservico

import (
	"context"

	"github.com/muriloperosa/soat-architecture/internal/domain/query"
)

// OrdemServicoRepository persiste e consulta Ordens de Serviço.
type OrdemServicoRepository interface {
	Salvar(ctx context.Context, ordemServico *OrdemServico) error
	BuscarPorID(ctx context.Context, id uint64) (*OrdemServico, error)
	BuscarPorNumero(ctx context.Context, numero string) (*OrdemServico, error)
	Listar(ctx context.Context, params query.Params) (query.Page[*OrdemServico], error)
	Atualizar(ctx context.Context, ordemServico *OrdemServico) error
	BuscarPorIDComBloqueio(ctx context.Context, id uint64) (*OrdemServico, error)
}
