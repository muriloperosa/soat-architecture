package relatorio

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrPeriodoInicioMaiorOuIgualFim = shared.NewValidationError("data inicial deve ser anterior à data final")
	ErrPeriodoFimNoFuturo           = shared.NewValidationError("data final não pode ser uma data futura")
	ErrPeriodoInicioAntesDoLimite   = shared.NewValidationError("data inicial não pode ser anterior a 01/01/2026")
	ErrTransicaoStatusIguais        = shared.NewValidationError("status de origem e destino não podem ser iguais")
	ErrTransicaoStatusSemCaminho    = shared.NewValidationError("não existe caminho válido entre os status informados")
)
