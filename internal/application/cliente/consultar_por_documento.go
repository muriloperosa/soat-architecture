package cliente

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type ConsultarClientePorDocumentoUseCase struct {
	repository domain.Repository
}

func NewConsultarClientePorDocumentoUseCase(
	repository domain.Repository,
) *ConsultarClientePorDocumentoUseCase {
	return &ConsultarClientePorDocumentoUseCase{repository: repository}
}

func (uc *ConsultarClientePorDocumentoUseCase) Executar(
	ctx context.Context,
	documento string,
) (ClienteOutput, error) {
	cliente, err := uc.repository.BuscarPorDocumento(ctx, documento)
	if err != nil {
		return ClienteOutput{}, err
	}

	return toOutput(cliente), nil
}
