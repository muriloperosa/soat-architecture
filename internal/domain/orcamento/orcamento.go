// Package orcamento contém o Aggregate Root Orcamento e suas entidades
// internas ItemServico e ItemPeca.
package orcamento

import (
	"strings"
	"time"
)

// tamanhoMaximoObservacoes espelha o VARCHAR(500) da coluna
// orcamentos.observacoes (migration 000008).
const tamanhoMaximoObservacoes = 500

// Orcamento representa a estimativa de custo de uma Ordem de Serviço. Uma OS
// possui no máximo um Orçamento. O Orçamento não possui status próprio: a
// aprovação/rejeição pertence ao status da OS.
type Orcamento struct {
	id                uint64
	ordemServicoID    uint64
	itensServico      []ItemServico
	itensPeca         []ItemPeca
	valorItemServicos float64
	valorItemPecas    float64
	valorTotal        float64
	observacoes       string
	criadoPor         uint64
	criadoEm          time.Time
	atualizadoEm      time.Time
}

// NewOrcamento cria um Orçamento vazio para uma Ordem de Serviço.
func NewOrcamento(ordemServicoID uint64, observacoes string, criadoPor uint64) (*Orcamento, error) {
	if ordemServicoID == 0 {
		return nil, ErrOrdemServicoObrigatoria
	}
	if criadoPor == 0 {
		return nil, ErrCriadoPorObrigatorio
	}

	observacoes = strings.TrimSpace(observacoes)
	if len(observacoes) > tamanhoMaximoObservacoes {
		return nil, ErrObservacoesInvalidas
	}

	agora := time.Now()
	return &Orcamento{
		ordemServicoID: ordemServicoID,
		itensServico:   []ItemServico{},
		itensPeca:      []ItemPeca{},
		observacoes:    observacoes,
		criadoPor:      criadoPor,
		criadoEm:       agora,
		atualizadoEm:   agora,
	}, nil
}

// ReidratarOrcamento recompõe um Orçamento a partir dos dados persistidos;
// não reaplica validação de negócio. Usado só pelo mapper de persistência.
func ReidratarOrcamento(
	id, ordemServicoID uint64,
	itensServico []ItemServico,
	itensPeca []ItemPeca,
	valorItemServicos, valorItemPecas, valorTotal float64,
	observacoes string,
	criadoPor uint64,
	criadoEm, atualizadoEm time.Time,
) *Orcamento {
	servicos := append([]ItemServico(nil), itensServico...)
	pecas := append([]ItemPeca(nil), itensPeca...)

	return &Orcamento{
		id:                id,
		ordemServicoID:    ordemServicoID,
		itensServico:      servicos,
		itensPeca:         pecas,
		valorItemServicos: valorItemServicos,
		valorItemPecas:    valorItemPecas,
		valorTotal:        valorTotal,
		observacoes:       observacoes,
		criadoPor:         criadoPor,
		criadoEm:          criadoEm,
		atualizadoEm:      atualizadoEm,
	}
}

// AtribuirID registra a identidade gerada pela persistência no Orçamento.
func (o *Orcamento) AtribuirID(id uint64) {
	o.id = id
}

// AdicionarItemServico inclui um serviço no orçamento com o valor e tempo
// estimado vigentes no momento da inclusão, e recalcula os totais.
func (o *Orcamento) AdicionarItemServico(servicoID uint64, quantidade int, valor float64, tempoEstimadoMinutos int) error {
	item, err := NewItemServico(servicoID, quantidade, valor, tempoEstimadoMinutos)
	if err != nil {
		return err
	}

	o.itensServico = append(o.itensServico, item)
	o.atualizadoEm = time.Now()
	o.CalcularTotal()
	return nil
}

// AdicionarItemPeca inclui uma peça no orçamento com a descrição e valor
// vigentes no momento da inclusão, e recalcula os totais.
func (o *Orcamento) AdicionarItemPeca(pecaID uint64, descricao string, quantidade int, valor float64) error {
	item, err := NewItemPeca(pecaID, descricao, quantidade, valor)
	if err != nil {
		return err
	}

	o.itensPeca = append(o.itensPeca, item)
	o.atualizadoEm = time.Now()
	o.CalcularTotal()
	return nil
}

// RemoverItemServico remove um item de serviço pelo ID e recalcula os totais.
func (o *Orcamento) RemoverItemServico(itemID uint64) error {
	for i, item := range o.itensServico {
		if item.ID() == itemID {
			o.itensServico = append(o.itensServico[:i], o.itensServico[i+1:]...)
			o.atualizadoEm = time.Now()
			o.CalcularTotal()
			return nil
		}
	}
	return ErrItemServicoNaoEncontrado
}

// RemoverItemPeca remove um item de peça pelo ID e recalcula os totais.
func (o *Orcamento) RemoverItemPeca(itemID uint64) error {
	for i, item := range o.itensPeca {
		if item.ID() == itemID {
			o.itensPeca = append(o.itensPeca[:i], o.itensPeca[i+1:]...)
			o.atualizadoEm = time.Now()
			o.CalcularTotal()
			return nil
		}
	}
	return ErrItemPecaNaoEncontrado
}

// CalcularTotal recalcula valorItemServicos, valorItemPecas e valorTotal a
// partir dos itens correntes, e devolve o valorTotal resultante.
func (o *Orcamento) CalcularTotal() float64 {
	var totalServicos, totalPecas float64
	for _, item := range o.itensServico {
		totalServicos += item.CalcularSubtotal()
	}
	for _, item := range o.itensPeca {
		totalPecas += item.CalcularSubtotal()
	}

	o.valorItemServicos = totalServicos
	o.valorItemPecas = totalPecas
	o.valorTotal = totalServicos + totalPecas
	return o.valorTotal
}

// ValidarParaEnvio garante que o orçamento possui conteúdo antes de ser
// enviado para aprovação. O orçamento não possui status próprio; seu ciclo
// de aprovação é controlado pela Ordem de Serviço.
func (o *Orcamento) ValidarParaEnvio() error {
	if len(o.itensServico) == 0 && len(o.itensPeca) == 0 {
		return ErrOrcamentoVazio
	}

	o.CalcularTotal()
	return nil
}

func (o *Orcamento) ID() uint64                 { return o.id }
func (o *Orcamento) OrdemServicoID() uint64     { return o.ordemServicoID }
func (o *Orcamento) ValorItemServicos() float64 { return o.valorItemServicos }
func (o *Orcamento) ValorItemPecas() float64    { return o.valorItemPecas }
func (o *Orcamento) ValorTotal() float64        { return o.valorTotal }
func (o *Orcamento) Observacoes() string        { return o.observacoes }
func (o *Orcamento) CriadoPor() uint64          { return o.criadoPor }
func (o *Orcamento) CriadoEm() time.Time        { return o.criadoEm }
func (o *Orcamento) AtualizadoEm() time.Time    { return o.atualizadoEm }

// ItensServico devolve uma cópia para preservar o encapsulamento do agregado.
func (o *Orcamento) ItensServico() []ItemServico {
	return append([]ItemServico(nil), o.itensServico...)
}

// ItensPeca devolve uma cópia para preservar o encapsulamento do agregado.
func (o *Orcamento) ItensPeca() []ItemPeca {
	return append([]ItemPeca(nil), o.itensPeca...)
}
