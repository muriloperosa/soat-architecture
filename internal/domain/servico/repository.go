package servico

import "context"

// ServicoRepository persiste e consulta o catálogo de serviços da oficina.
// BuscarPorID retorna ErrServicoNaoEncontrado (sentinel, não
// gorm.ErrRecordNotFound cru) quando o serviço não existe.
type ServicoRepository interface {
	Salvar(ctx context.Context, s *Servico) error
	BuscarPorID(ctx context.Context, id uint64) (*Servico, error)
	Listar(ctx context.Context) ([]*Servico, error)
	Atualizar(ctx context.Context, s *Servico) error
}
