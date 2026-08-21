package peca

import "context"

// Repository persiste e consulta peças. BuscarPorID e BuscarPorCodigo
// retornam ErrPecaNaoEncontrada (sentinel, não gorm.ErrRecordNotFound cru)
// quando a peça não existe.
type Repository interface {
	Salvar(ctx context.Context, peca *Peca) error
	BuscarPorID(ctx context.Context, id uint64) (*Peca, error)
	BuscarPorCodigo(ctx context.Context, codigo string) (*Peca, error)
	Atualizar(ctx context.Context, peca *Peca) error
}
