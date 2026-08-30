package orcamento

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

// ItemServico é um serviço incluído no Orçamento. O valor e o tempo estimado
// são preservados no momento da inclusão: alterações futuras no cadastro do
// Serviço não alteram itens já incluídos em orçamentos existentes.
type ItemServico struct {
	id            uint64
	orcamentoID   uint64
	servicoID     uint64
	quantidade    int
	valor         float64
	tempoEstimado shared.DuracaoEstimada
}

// NewItemServico cria um item de serviço a partir dos dados vigentes do
// catálogo no momento da inclusão (valor e tempo estimado já resolvidos).
func NewItemServico(servicoID uint64, quantidade int, valor float64, tempoEstimadoMinutos int) (ItemServico, error) {
	if servicoID == 0 {
		return ItemServico{}, ErrServicoObrigatorio
	}
	if quantidade <= 0 {
		return ItemServico{}, ErrQuantidadeInvalida
	}
	if valor < 0 {
		return ItemServico{}, ErrValorInvalido
	}

	tempoEstimado, err := shared.NewDuracaoEstimada(tempoEstimadoMinutos)
	if err != nil {
		return ItemServico{}, err
	}

	return ItemServico{
		servicoID:     servicoID,
		quantidade:    quantidade,
		valor:         valor,
		tempoEstimado: tempoEstimado,
	}, nil
}

// ReidratarItemServico recompõe um item a partir dos dados persistidos; não
// reaplica validação de negócio. Usado só pelo mapper de persistência.
func ReidratarItemServico(
	id, orcamentoID, servicoID uint64,
	quantidade int,
	valor float64,
	tempoEstimado shared.DuracaoEstimada,
) ItemServico {
	return ItemServico{
		id:            id,
		orcamentoID:   orcamentoID,
		servicoID:     servicoID,
		quantidade:    quantidade,
		valor:         valor,
		tempoEstimado: tempoEstimado,
	}
}

// CalcularSubtotal aplica subtotal = valor × quantidade.
func (i ItemServico) CalcularSubtotal() float64 {
	return i.valor * float64(i.quantidade)
}

func (i ItemServico) ID() uint64          { return i.id }
func (i ItemServico) OrcamentoID() uint64 { return i.orcamentoID }
func (i ItemServico) ServicoID() uint64   { return i.servicoID }
func (i ItemServico) Quantidade() int     { return i.quantidade }
func (i ItemServico) Valor() float64      { return i.valor }

func (i ItemServico) TempoEstimado() shared.DuracaoEstimada { return i.tempoEstimado }
