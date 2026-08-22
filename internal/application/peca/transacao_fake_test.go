package peca

import "context"

// transacaoFake executa fn direto, sem transação de verdade — os testes de
// use case usam mocks de Repository, então não há conexão real com quem
// coordenar. Serve só pra satisfazer shared.TransactionRunner nos testes.
type transacaoFake struct{}

func (transacaoFake) Executar(ctx context.Context, fn func(ctx context.Context) error) error {
	return fn(ctx)
}
