package servico

import (
	"context"
	"errors"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// InativarServicoUseCase bloqueia um serviço do catálogo para novas OS.
type InativarServicoUseCase struct {
	repo domainservico.ServicoRepository
}

func NewInativarServicoUseCase(repo domainservico.ServicoRepository) *InativarServicoUseCase {
	return &InativarServicoUseCase{repo: repo}
}

func (uc *InativarServicoUseCase) Executar(ctx context.Context, id uint64) error {
	s, err := uc.repo.BuscarPorID(ctx, id)
	if err != nil {
		if errors.Is(err, domainservico.ErrServicoNaoEncontrado) {
			return err
		}
		return shared.NewInternalError("erro ao buscar serviço", err)
	}

	s.Inativar()

	if err := uc.repo.Atualizar(ctx, s); err != nil {
		return shared.NewInternalError("erro ao inativar serviço", err)
	}
	return nil
}
