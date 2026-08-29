package relatorio

import "time"

// ConsultarTransicaoStatusInput reúne os parâmetros do relatório de
// transição de status de Ordem de Serviço.
type ConsultarTransicaoStatusInput struct {
	DataInicio time.Time
	DataFim    time.Time
	DeStatus   string
	ParaStatus string
}

// ConsultarTransicaoStatusOutput é o resultado agregado do relatório.
// As durações ficam sem unidade fixa de propósito — formatar na unidade
// pedida (h/m/s) é responsabilidade da camada HTTP.
type ConsultarTransicaoStatusOutput struct {
	TotalOrdensServico int
	TempoMedio         time.Duration
	TempoMinimo        time.Duration
	TempoMaximo        time.Duration
}
