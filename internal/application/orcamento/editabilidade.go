package orcamento

import (
	"context"
	"errors"

	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

func validarOrcamentoEditavel(
	ctx context.Context,
	repository domainordemservico.OrdemServicoRepository,
	ordemServicoID uint64,
) error {
	os, err := repository.BuscarPorID(ctx, ordemServicoID)
	if err != nil {
		if errors.Is(err, domainordemservico.ErrOrdemServicoNaoEncontrada) {
			return err
		}
		return shared.NewInternalError("erro ao buscar ordem de serviço", err)
	}

	switch os.Status() {
	case domainordemservico.StatusEmDiagnostico, domainordemservico.StatusRejeitada:
		return nil
	default:
		return domainorcamento.ErrOrcamentoImutavel
	}
}
