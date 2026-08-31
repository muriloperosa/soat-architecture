package orcamento

import "strings"

// tamanhoMaximoDescricao espelha o VARCHAR(500) da coluna
// orcamentos_pecas.descricao (migration 000010).
const tamanhoMaximoDescricao = 500

// ItemPeca é uma peça incluída no Orçamento. Descrição e valor são
// preservados no momento da inclusão: alterações futuras no cadastro da
// Peça não alteram itens já incluídos em orçamentos existentes. Não
// representa reserva de estoque.
type ItemPeca struct {
	id          uint64
	orcamentoID uint64
	pecaID      uint64
	descricao   string
	quantidade  int
	valor       float64
}

// NewItemPeca cria um item de peça a partir dos dados vigentes do cadastro
// no momento da inclusão (descrição e valor já resolvidos).
func NewItemPeca(pecaID uint64, descricao string, quantidade int, valor float64) (ItemPeca, error) {
	if pecaID == 0 {
		return ItemPeca{}, ErrPecaObrigatoria
	}

	descricao = strings.TrimSpace(descricao)
	if descricao == "" {
		return ItemPeca{}, ErrDescricaoObrigatoria
	}
	if len(descricao) > tamanhoMaximoDescricao {
		return ItemPeca{}, ErrDescricaoInvalida
	}

	if quantidade <= 0 {
		return ItemPeca{}, ErrQuantidadeInvalida
	}
	if valor < 0 {
		return ItemPeca{}, ErrValorInvalido
	}

	return ItemPeca{
		pecaID:     pecaID,
		descricao:  descricao,
		quantidade: quantidade,
		valor:      valor,
	}, nil
}

// ReidratarItemPeca recompõe um item a partir dos dados persistidos; não
// reaplica validação de negócio. Usado só pelo mapper de persistência.
func ReidratarItemPeca(id, orcamentoID, pecaID uint64, descricao string, quantidade int, valor float64) ItemPeca {
	return ItemPeca{
		id:          id,
		orcamentoID: orcamentoID,
		pecaID:      pecaID,
		descricao:   descricao,
		quantidade:  quantidade,
		valor:       valor,
	}
}

// AlterarQuantidade altera a quantidade do item do orçamento. A reserva de
// estoque não é atualizada aqui: quando um orçamento já aprovado é alterado,
// a application invalida as reservas antigas e exige nova aprovação.
func (i *ItemPeca) AlterarQuantidade(quantidade int) error {
	if quantidade <= 0 {
		return ErrQuantidadeInvalida
	}

	i.quantidade = quantidade
	return nil
}

// CalcularSubtotal aplica subtotal = valor × quantidade.
func (i ItemPeca) CalcularSubtotal() float64 {
	return i.valor * float64(i.quantidade)
}

func (i ItemPeca) ID() uint64          { return i.id }
func (i ItemPeca) OrcamentoID() uint64 { return i.orcamentoID }
func (i ItemPeca) PecaID() uint64      { return i.pecaID }
func (i ItemPeca) Descricao() string   { return i.descricao }
func (i ItemPeca) Quantidade() int     { return i.quantidade }
func (i ItemPeca) Valor() float64      { return i.valor }
