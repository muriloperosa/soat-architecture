package peca

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

type CadastrarPecaUseCase struct {
	repository domain.Repository
}

func NewCadastrarPecaUseCase(repository domain.Repository) *CadastrarPecaUseCase {
	return &CadastrarPecaUseCase{repository: repository}
}

func (uc *CadastrarPecaUseCase) Executar(ctx context.Context, input CadastrarPecaInput) (PecaOutput, error) {
	peca, err := domain.NewPeca(
		input.Nome,
		input.Marca,
		input.Descricao,
		input.Preco,
		input.QuantidadeEmEstoque,
		input.EstoqueMinimo,
		input.CriadoPor,
	)
	if err != nil {
		return PecaOutput{}, err
	}

	if err := uc.repository.Salvar(ctx, peca); err != nil {
		return PecaOutput{}, err
	}

	return toOutput(peca), nil
}
