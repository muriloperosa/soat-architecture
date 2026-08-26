package shared

var ErrDuracaoEstimadaInvalida = NewValidationError("tempo estimado deve ser maior que zero")

// DuracaoEstimada é o VO de tempo estimado de um serviço, em minutos.
// Reusado por Serviço (catálogo) e ItemServico (orçamento).
type DuracaoEstimada struct {
	minutos int
}

// NewDuracaoEstimada cria o VO a partir de minutos. Rejeita zero e negativo
// (o CHECK do banco também exige tempo_estimado_minutos > 0).
func NewDuracaoEstimada(minutos int) (DuracaoEstimada, error) {
	if minutos <= 0 {
		return DuracaoEstimada{}, ErrDuracaoEstimadaInvalida
	}
	return DuracaoEstimada{minutos: minutos}, nil
}

// RestaurarDuracaoEstimada reidrata o VO a partir de minutos já persistidos.
// Não revalida. Usado só por mappers de persistência.
func RestaurarDuracaoEstimada(minutos int) DuracaoEstimada {
	return DuracaoEstimada{minutos: minutos}
}

// Minutos retorna a duração em minutos.
func (d DuracaoEstimada) Minutos() int {
	return d.minutos
}

// Horas retorna a duração em horas (fração permitida, ex.: 90 min = 1.5).
func (d DuracaoEstimada) Horas() float64 {
	return float64(d.minutos) / 60
}
