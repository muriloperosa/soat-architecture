package orcamento

import (
	"context"
	"errors"

	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// AdicionarPecaOrcamentoUseCase inclui uma peça do estoque no orçamento,
// copiando a descrição e o valor vigentes no momento da inclusão para
// preservar o histórico do orçamento.
type AdicionarPecaOrcamentoUseCase struct {
	orcamentoRepository    domainorcamento.OrcamentoRepository
	pecaRepository         domainpeca.Repository
	ordemServicoRepository domainordemservico.OrdemServicoRepository
}

func NewAdicionarPecaOrcamentoUseCase(
	orcamentoRepository domainorcamento.OrcamentoRepository,
	pecaRepository domainpeca.Repository,
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
) *AdicionarPecaOrcamentoUseCase {
	return &AdicionarPecaOrcamentoUseCase{
		orcamentoRepository:    orcamentoRepository,
		pecaRepository:         pecaRepository,
		ordemServicoRepository: ordemServicoRepository,
	}
}

func (uc *AdicionarPecaOrcamentoUseCase) Executar(ctx context.Context, input AdicionarPecaOrcamentoInput) (OrcamentoOutput, error) {
	if err := validarOrcamentoEditavel(ctx, uc.ordemServicoRepository, input.OrdemServicoID); err != nil {
		return OrcamentoOutput{}, err
	}

	orcamento, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainorcamento.ErrOrcamentoNaoEncontrado) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar orçamento", err)
	}

	peca, err := uc.pecaRepository.BuscarPorID(ctx, input.PecaID)
	if err != nil {
		if errors.Is(err, domainpeca.ErrPecaNaoEncontrada) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar peça", err)
	}

	if err := orcamento.AdicionarItemPeca(
		peca.ID(),
		peca.Descricao(),
		input.Quantidade,
		peca.Preco(),
	); err != nil {
		return OrcamentoOutput{}, err
	}

	if err := uc.orcamentoRepository.Atualizar(ctx, orcamento); err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao atualizar orçamento", err)
	}

	atualizado, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar orçamento atualizado", err)
	}

	return toOutput(atualizado), nil
}
