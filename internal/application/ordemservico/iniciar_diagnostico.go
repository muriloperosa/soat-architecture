package ordemservico

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

type IniciarDiagnosticoUseCase struct {
	repository  domain.OrdemServicoRepository
	notificador *notificacao.NotificadorStatus
}

func NewIniciarDiagnosticoUseCase(repository domain.OrdemServicoRepository, notificador *notificacao.NotificadorStatus) *IniciarDiagnosticoUseCase {
	return &IniciarDiagnosticoUseCase{repository: repository, notificador: notificador}
}

func (uc *IniciarDiagnosticoUseCase) Executar(
	ctx context.Context,
	input IniciarDiagnosticoInput,
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
	if err := os.IniciarDiagnostico(input.UsuarioID); err != nil {
		return OrdemServicoOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, os); err != nil {
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao iniciar diagnóstico da ordem de serviço", err)
	}

	if err := uc.notificador.Notificar(ctx, os, statusAnterior); err != nil {
		return OrdemServicoOutput{}, err
	}

	return toOutput(os), nil
}
