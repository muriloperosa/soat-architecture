package orcamento

import (
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

func validarCicloAprovacao(os *domainordemservico.OrdemServico, approvalHistoryID uint64) error {
	if approvalHistoryID == 0 {
		return nil
	}

	currentID, ok := os.IDHistoricoAprovacaoAtual()
	if !ok || currentID != approvalHistoryID {
		return shared.NewConflictError("este link pertence a um envio anterior do orçamento; use o link mais recente")
	}
	return nil
}
