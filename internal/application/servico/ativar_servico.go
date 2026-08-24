package servico

import (
	"context"
	"errors"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AtivarServicoUseCase reabilita um serviço do catálogo para uso em OS.
type AtivarServicoUseCase struct {
	repo domainservico.ServicoRepository
}

func NewAtivarServicoUseCase(repo domainservico.ServicoRepository) *AtivarServicoUseCase {
	return &AtivarServicoUseCase{repo: repo}
}

func (uc *AtivarServicoUseCase) Executar(ctx context.Context, id uint64) error {
	s, err := uc.repo.BuscarPorID(ctx, id)
	if err != nil {
		if errors.Is(err, domainservico.ErrServicoNaoEncontrado) {
			return err
		}
		return shared.NewInternalError("erro ao buscar serviço", err)
	}

	s.Ativar()

	if err := uc.repo.Atualizar(ctx, s); err != nil {
		return shared.NewInternalError("erro ao ativar serviço", err)
	}
	return nil
}
