package ordemservico

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

const motivoInicioDiagnostico = "Diagnóstico iniciado"

type IniciarDiagnosticoUseCase struct {
	repository domain.OrdemServicoRepository
}

func NewIniciarDiagnosticoUseCase(repository domain.OrdemServicoRepository) *IniciarDiagnosticoUseCase {
	return &IniciarDiagnosticoUseCase{repository: repository}
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

	if err := os.IniciarDiagnostico(input.UsuarioID, motivoInicioDiagnostico); err != nil {
		return OrdemServicoOutput{}, err
	}

	if err := uc.repository.Atualizar(ctx, os); err != nil {
		return OrdemServicoOutput{}, shared.NewInternalError("erro ao iniciar diagnóstico da ordem de serviço", err)
	}

	return toOutput(os), nil
}
