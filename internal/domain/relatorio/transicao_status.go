package relatorio

import (
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
)

// CalcularTransicaoStatusParams reúne os parâmetros de consulta do relatório
// de transição de status de Ordem de Serviço.
type CalcularTransicaoStatusParams struct {
	FromStatus ordemservico.StatusOrdemServico
	ToStatus   ordemservico.StatusOrdemServico
	Periodo    Periodo
}

// TransicaoStatusResultado é o resultado agregado do relatório: quantas OS
// completaram a transição no período e a duração média/mínima/máxima.
type TransicaoStatusResultado struct {
	TotalOrdens   int
	DuracaoMedia  time.Duration
	DuracaoMinima time.Duration
	DuracaoMaxima time.Duration
}
