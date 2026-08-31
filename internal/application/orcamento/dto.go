package orcamento

import (
	domain "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
)

type GerarOrcamentoInput struct {
	OrdemServicoID uint64
	Observacoes    string
	UsuarioID      uint64
}

type AdicionarServicoOrcamentoInput struct {
	OrdemServicoID uint64
	ServicoID      uint64
	Quantidade     int
}

type AdicionarPecaOrcamentoInput struct {
	OrdemServicoID uint64
	PecaID         uint64
	Quantidade     int
}

type AlterarQuantidadePecaOrcamentoInput struct {
	OrdemServicoID uint64
	ItemPecaID     uint64
	Quantidade     int
	UsuarioID      uint64
}

type RemoverServicoOrcamentoInput struct {
	OrdemServicoID uint64
	ItemServicoID  uint64
}

type RemoverPecaOrcamentoInput struct {
	OrdemServicoID uint64
	ItemPecaID     uint64
}

type EnviarOrcamentoParaAprovacaoInput struct {
	OrdemServicoID uint64
	UsuarioID      uint64
}

type AprovarOrcamentoInput struct {
	OrdemServicoID uint64
	ClienteID      uint64
}

type RejeitarOrcamentoInput struct {
	OrdemServicoID uint64
	ClienteID      uint64
	Motivo         string
}

type ItemServicoOutput struct {
	ID                   uint64
	ServicoID            uint64
	Quantidade           int
	Valor                float64
	TempoEstimadoMinutos int
	Subtotal             float64
}

type ItemPecaOutput struct {
	ID         uint64
	PecaID     uint64
	Descricao  string
	Quantidade int
	Valor      float64
	Subtotal   float64
}

type OrcamentoOutput struct {
	ID                uint64
	OrdemServicoID    uint64
	ValorItemServicos float64
	ValorItemPecas    float64
	ValorTotal        float64
	Observacoes       string
	ItensServico      []ItemServicoOutput
	ItensPeca         []ItemPecaOutput
}

func toOutput(o *domain.Orcamento) OrcamentoOutput {
	itensServico := make([]ItemServicoOutput, 0, len(o.ItensServico()))
	for _, item := range o.ItensServico() {
		itensServico = append(itensServico, ItemServicoOutput{
			ID:                   item.ID(),
			ServicoID:            item.ServicoID(),
			Quantidade:           item.Quantidade(),
			Valor:                item.Valor(),
			TempoEstimadoMinutos: item.TempoEstimado().Minutos(),
			Subtotal:             item.CalcularSubtotal(),
		})
	}

	itensPeca := make([]ItemPecaOutput, 0, len(o.ItensPeca()))
	for _, item := range o.ItensPeca() {
		itensPeca = append(itensPeca, ItemPecaOutput{
			ID:         item.ID(),
			PecaID:     item.PecaID(),
			Descricao:  item.Descricao(),
			Quantidade: item.Quantidade(),
			Valor:      item.Valor(),
			Subtotal:   item.CalcularSubtotal(),
		})
	}

	return OrcamentoOutput{
		ID:                o.ID(),
		OrdemServicoID:    o.OrdemServicoID(),
		ValorItemServicos: o.ValorItemServicos(),
		ValorItemPecas:    o.ValorItemPecas(),
		ValorTotal:        o.ValorTotal(),
		Observacoes:       o.Observacoes(),
		ItensServico:      itensServico,
		ItensPeca:         itensPeca,
	}
}

// FluxoOrcamentoOutput representa o resultado das decisões que alteram o
// status da Ordem de Serviço.
type FluxoOrcamentoOutput struct {
	OrdemServicoID uint64
	Status         string
}

func toFluxoOutput(os *domainordemservico.OrdemServico) FluxoOrcamentoOutput {
	return FluxoOrcamentoOutput{
		OrdemServicoID: os.ID(),
		Status:         string(os.Status()),
	}
}
