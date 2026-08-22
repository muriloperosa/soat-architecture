package mysql

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type txContextKey struct{}

// DBFromContext retorna a transação injetada no ctx por TransactionRunner,
// se houver; senão retorna db. Repositórios devem chamar isso em vez de
// usar r.db direto, pra participar de uma transação cross-repository
// quando o use case pedir (ex. Repository.Salvar(ctx, ...) usa
// DBFromContext(ctx, r.db) ao montar a query).
func DBFromContext(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txContextKey{}).(*gorm.DB); ok {
		return tx
	}
	return db
}

// ComBloqueio adiciona SELECT ... FOR UPDATE à query. Só tem efeito de
// verdade dentro de uma transação (fora dela, o lock é liberado no fim do
// autocommit da própria SELECT) — use sempre com db vindo de
// DBFromContext dentro de um shared.TransactionRunner.Executar.
func ComBloqueio(db *gorm.DB) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"})
}

// TransactionRunner implementa [shared.TransactionRunner] sobre GORM/MySQL.
type TransactionRunner struct {
	db *gorm.DB
}

var _ shared.TransactionRunner = (*TransactionRunner)(nil)

func NewTransactionRunner(db *gorm.DB) *TransactionRunner {
	return &TransactionRunner{db: db}
}

// Executar implements [shared.TransactionRunner].
func (t *TransactionRunner) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	return t.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(context.WithValue(ctx, txContextKey{}, tx))
	})
}
