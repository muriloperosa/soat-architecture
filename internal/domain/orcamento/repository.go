package orcamento

import "context"

// OrcamentoRepository persiste e consulta Orçamentos. BuscarPorOrdemServicoID
// retorna ErrOrcamentoNaoEncontrado (sentinel, não gorm.ErrRecordNotFound
// cru) quando a OS não possui orçamento.
type OrcamentoRepository interface {
	Salvar(ctx context.Context, orcamento *Orcamento) error
	BuscarPorOrdemServicoID(ctx context.Context, ordemServicoID uint64) (*Orcamento, error)
	BuscarPorOrdensServicoIDs(ctx context.Context, ordensServicoIDs []uint64) ([]*Orcamento, error)
	Atualizar(ctx context.Context, orcamento *Orcamento) error
}
