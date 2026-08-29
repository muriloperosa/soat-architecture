package relatorio

import "context"

// RelatorioTransicaoStatusRepository calcula, a partir do histórico de
// status persistido, o total e a duração da transição from->to no período.
type RelatorioTransicaoStatusRepository interface {
	CalcularTransicaoStatus(ctx context.Context, params CalcularTransicaoStatusParams) (TransicaoStatusResultado, error)
}
