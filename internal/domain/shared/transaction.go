package shared

import "context"

// TransactionRunner executa fn dentro de uma transação atômica. Erro
// retornado por fn causa rollback; nil causa commit. Usado por use cases
// que precisam coordenar mais de um Repository (de agregados diferentes)
// numa única operação atômica, como reservar estoque de uma Peça.
type TransactionRunner interface {
	Executar(ctx context.Context, fn func(ctx context.Context) error) error
}
