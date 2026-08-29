package relatorio

import "time"

var dataMinimaPeriodo = time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC)

// Periodo é o VO de intervalo de datas usado pelos relatórios. Nasce válido
// ou não nasce: início antes do fim, fim não pode ser futuro, início não
// pode ser anterior ao corte de dados confiáveis (01/01/2026).
type Periodo struct {
	inicio time.Time
	fim    time.Time
}

// NewPeriodo valida e cria o VO Periodo.
func NewPeriodo(inicio, fim time.Time) (Periodo, error) {
	if !inicio.Before(fim) {
		return Periodo{}, ErrPeriodoInicioMaiorOuIgualFim
	}
	if fim.After(time.Now()) {
		return Periodo{}, ErrPeriodoFimNoFuturo
	}
	if inicio.Before(dataMinimaPeriodo) {
		return Periodo{}, ErrPeriodoInicioAntesDoLimite
	}

	return Periodo{inicio: inicio, fim: fim}, nil
}

// Inicio retorna o início do período.
func (p Periodo) Inicio() time.Time {
	return p.inicio
}

// Fim retorna o fim do período.
func (p Periodo) Fim() time.Time {
	return p.fim
}
