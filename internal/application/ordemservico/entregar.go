package ordemservico

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type EntregarOrdemServicoUseCase struct {
	repository  domain.OrdemServicoRepository
	notificador *notificacao.NotificadorStatus
}

func NewEntregarOrdemServicoUseCase(repository domain.OrdemServicoRepository, notificador *notificacao.NotificadorStatus) *EntregarOrdemServicoUseCase {
	return &EntregarOrdemServicoUseCase{repository: repository, notificador: notificador}
}

func (uc *EntregarOrdemServicoUseCase) Executar(
	ctx context.Context,
	input EntregarOrdemServicoInput,
) (OrdemServicoOutput, error) {
	os, err := uc.repository.BuscarPorID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domain.ErrOrdemServicoNaoEncontrada) {
			return OrdemServicoOutput{}, err
		}
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao buscar ordem de serviço", err)
	}
	if os == nil {
		return OrdemServicoOutput{}, domain.ErrOrdemServicoNaoEncontrada
	}

	statusAnterior := os.Status()
	if err := os.Entregar(input.UsuarioID); err != nil {
		return OrdemServicoOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, os); err != nil {
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao entregar ordem de serviço", err)
	}

	uc.notificador.Notificar(ctx, os, statusAnterior)

	return toOutput(os), nil
}
