package servico

import (
	"context"
	"errors"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// BuscarServicoUseCase consulta um serviço do catálogo pelo ID.
type BuscarServicoUseCase struct {
	repo domainservico.ServicoRepository
}

func NewBuscarServicoUseCase(repo domainservico.ServicoRepository) *BuscarServicoUseCase {
	return &BuscarServicoUseCase{repo: repo}
}

func (uc *BuscarServicoUseCase) Executar(ctx context.Context, id uint64) (ServicoOutput, error) {
	s, err := uc.repo.BuscarPorID(ctx, id)
	if err != nil {
		if errors.Is(err, domainservico.ErrServicoNaoEncontrado) {
			return ServicoOutput{}, err
		}
		return ServicoOutput{}, shared.NewInternalError("erro ao buscar serviço", err)
	}
	return toOutput(s), nil
}
