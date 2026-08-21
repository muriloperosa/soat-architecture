package servico

import (
	"context"
	"errors"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AtualizarServicoUseCase troca nome, descrição, preço base e tempo estimado
// de um serviço existente. Não altera criadoPor nem ativo.
type AtualizarServicoUseCase struct {
	repo domainservico.ServicoRepository
}

func NewAtualizarServicoUseCase(repo domainservico.ServicoRepository) *AtualizarServicoUseCase {
	return &AtualizarServicoUseCase{repo: repo}
}

func (uc *AtualizarServicoUseCase) Executar(ctx context.Context, input AtualizarServicoInput) (ServicoOutput, error) {
	s, err := uc.repo.BuscarPorID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainservico.ErrServicoNaoEncontrado) {
			return ServicoOutput{}, err
		}
		return ServicoOutput{}, shared.NewInternalError("erro ao buscar serviço", err)
	}

	if err := s.Atualizar(input.Nome, input.Descricao, input.PrecoBase, input.TempoEstimadoMinutos); err != nil {
		return ServicoOutput{}, err
	}

	if err := uc.repo.Atualizar(ctx, s); err != nil {
		return ServicoOutput{}, shared.NewInternalError("erro ao atualizar serviço", err)
	}

	return toOutput(s), nil
}
