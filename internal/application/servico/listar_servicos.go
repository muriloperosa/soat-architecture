package servico

import (
	"context"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// ListarServicosUseCase devolve o catálogo completo de serviços.
type ListarServicosUseCase struct {
	repo domainservico.ServicoRepository
}

func NewListarServicosUseCase(repo domainservico.ServicoRepository) *ListarServicosUseCase {
	return &ListarServicosUseCase{repo: repo}
}

func (uc *ListarServicosUseCase) Executar(ctx context.Context) ([]ServicoOutput, error) {
	servicos, err := uc.repo.Listar(ctx)
	if err != nil {
		return nil, shared.NewInternalError("erro ao listar serviços", err)
	}
	return toOutputList(servicos), nil
}
