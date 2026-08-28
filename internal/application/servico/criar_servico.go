package servico

import (
	"context"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// CriarServicoUseCase cadastra um item no catálogo de serviços da oficina.
type CriarServicoUseCase struct {
	repo domainservico.ServicoRepository
}

func NewCriarServicoUseCase(repo domainservico.ServicoRepository) *CriarServicoUseCase {
	return &CriarServicoUseCase{repo: repo}
}

func (uc *CriarServicoUseCase) Executar(ctx context.Context, input CriarServicoInput) (ServicoOutput, error) {
	s, err := domainservico.NewServico(input.Nome, input.Descricao, input.PrecoBase, input.TempoEstimadoMinutos, input.CriadoPor)
	if err != nil {
		return ServicoOutput{}, err
	}

	if err := uc.repo.Salvar(ctx, s); err != nil {
		return ServicoOutput{}, shared.NewInternalError("erro ao salvar serviço", err)
	}

	return toOutput(s), nil
}
