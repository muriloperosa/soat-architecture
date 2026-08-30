package orcamento

import (
	domain "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

func toModel(o *domain.Orcamento) *OrcamentoModel {
	return &OrcamentoModel{
		ID:                o.ID(),
		OrdemServicoID:    o.OrdemServicoID(),
		ValorItemServicos: o.ValorItemServicos(),
		ValorItemPecas:    o.ValorItemPecas(),
		ValorTotal:        o.ValorTotal(),
		Observacoes:       o.Observacoes(),
		CriadoPor:         o.CriadoPor(),
		CriadoEm:          o.CriadoEm(),
		AtualizadoEm:      o.AtualizadoEm(),
	}
}

func toItemServicoModel(item domain.ItemServico, orcamentoID uint64) ItemServicoModel {
	return ItemServicoModel{
		ID:                   item.ID(),
		OrcamentoID:          orcamentoID,
		ServicoID:            item.ServicoID(),
		Quantidade:           item.Quantidade(),
		Valor:                item.Valor(),
		TempoEstimadoMinutos: item.TempoEstimado().Minutos(),
	}
}

func toItemPecaModel(item domain.ItemPeca, orcamentoID uint64) ItemPecaModel {
	return ItemPecaModel{
		ID:          item.ID(),
		OrcamentoID: orcamentoID,
		PecaID:      item.PecaID(),
		Descricao:   item.Descricao(),
		Quantidade:  item.Quantidade(),
		Valor:       item.Valor(),
	}
}

func toDomain(model OrcamentoModel) *domain.Orcamento {
	itensServico := make([]domain.ItemServico, 0, len(model.ItensServico))
	for _, itemModel := range model.ItensServico {
		itensServico = append(itensServico, domain.ReidratarItemServico(
			itemModel.ID,
			itemModel.OrcamentoID,
			itemModel.ServicoID,
			itemModel.Quantidade,
			itemModel.Valor,
			shared.RestaurarDuracaoEstimada(itemModel.TempoEstimadoMinutos),
		))
	}

	itensPeca := make([]domain.ItemPeca, 0, len(model.ItensPeca))
	for _, itemModel := range model.ItensPeca {
		itensPeca = append(itensPeca, domain.ReidratarItemPeca(
			itemModel.ID,
			itemModel.OrcamentoID,
			itemModel.PecaID,
			itemModel.Descricao,
			itemModel.Quantidade,
			itemModel.Valor,
		))
	}

	return domain.ReidratarOrcamento(
		model.ID,
		model.OrdemServicoID,
		itensServico,
		itensPeca,
		model.ValorItemServicos,
		model.ValorItemPecas,
		model.ValorTotal,
		model.Observacoes,
		model.CriadoPor,
		model.CriadoEm,
		model.AtualizadoEm,
	)
}
