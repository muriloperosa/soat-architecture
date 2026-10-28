package ordemservico

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type IniciarExecucaoUseCase struct {
	repository  domain.OrdemServicoRepository
	notificador *notificacao.NotificadorStatus
}

func NewIniciarExecucaoUseCase(repository domain.OrdemServicoRepository, notificador *notificacao.NotificadorStatus) *IniciarExecucaoUseCase {
	return &IniciarExecucaoUseCase{repository: repository, notificador: notificador}
}

func (uc *IniciarExecucaoUseCase) Executar(
	ctx context.Context,
	input IniciarExecucaoInput,
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
	if err := os.IniciarExecucao(input.UsuarioID); err != nil {
		return OrdemServicoOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, os); err != nil {
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao iniciar execução da ordem de serviço", err)
	}

	if err := uc.notificador.Notificar(ctx, os, statusAnterior); err != nil {
		return OrdemServicoOutput{}, err
	}

	return toOutput(os), nil
}
