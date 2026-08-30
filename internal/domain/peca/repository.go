package peca

import (
	"context"

	"github.com/muriloperosa/soat-architecture/internal/domain/query"
)

// Repository persiste e consulta peças. BuscarPorID e BuscarPorCodigo
// retornam ErrPecaNaoEncontrada (sentinel, não gorm.ErrRecordNotFound cru)
// quando a peça não existe.
type Repository interface {
	Salvar(ctx context.Context, peca *Peca) error
	BuscarPorID(ctx context.Context, id uint64) (*Peca, error)
	BuscarPorCodigo(ctx context.Context, codigo string) (*Peca, error)
	Atualizar(ctx context.Context, peca *Peca) error
	Listar(ctx context.Context, params query.Params) (query.Page[*Peca], error)

	// BuscarPorIDComBloqueio busca a peça travando a linha (SELECT ... FOR
	// UPDATE) até o fim da transação corrente. Só faz sentido chamado
	// dentro de um shared.TransactionRunner.Executar, pra serializar
	// operações concorrentes (ex. reservar estoque) que leem e decidem com
	// base no mesmo estado.
	BuscarPorIDComBloqueio(ctx context.Context, id uint64) (*Peca, error)
}
