package orcamento

import app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"

func toGerarInput(ordemServicoID, usuarioID uint64, request GerarOrcamentoRequest) app.GerarOrcamentoInput {
	return app.GerarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		Observacoes:    request.Observacoes,
		UsuarioID:      usuarioID,
	}
}

func toAdicionarServicoInput(ordemServicoID uint64, request AdicionarServicoOrcamentoRequest) app.AdicionarServicoOrcamentoInput {
	return app.AdicionarServicoOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ServicoID:      request.ServicoID,
		Quantidade:     request.Quantidade,
	}
}

func toAdicionarPecaInput(ordemServicoID uint64, request AdicionarPecaOrcamentoRequest) app.AdicionarPecaOrcamentoInput {
	return app.AdicionarPecaOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		PecaID:         request.PecaID,
		Quantidade:     request.Quantidade,
	}
}

func toRemoverServicoInput(ordemServicoID, itemServicoID uint64) app.RemoverServicoOrcamentoInput {
	return app.RemoverServicoOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ItemServicoID:  itemServicoID,
	}
}

func toRemoverPecaInput(ordemServicoID, itemPecaID uint64) app.RemoverPecaOrcamentoInput {
	return app.RemoverPecaOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ItemPecaID:     itemPecaID,
	}
}

func toResponse(output app.OrcamentoOutput) OrcamentoResponse {
	itensServico := make([]ItemServicoResponse, 0, len(output.ItensServico))
	for _, item := range output.ItensServico {
		itensServico = append(itensServico, ItemServicoResponse{
			ID:                   item.ID,
			ServicoID:            item.ServicoID,
			Quantidade:           item.Quantidade,
			Valor:                item.Valor,
			TempoEstimadoMinutos: item.TempoEstimadoMinutos,
			Subtotal:             item.Subtotal,
		})
	}

	itensPeca := make([]ItemPecaResponse, 0, len(output.ItensPeca))
	for _, item := range output.ItensPeca {
		itensPeca = append(itensPeca, ItemPecaResponse{
			ID:         item.ID,
			PecaID:     item.PecaID,
			Descricao:  item.Descricao,
			Quantidade: item.Quantidade,
			Valor:      item.Valor,
			Subtotal:   item.Subtotal,
		})
	}

	return OrcamentoResponse{
		ID:                output.ID,
		OrdemServicoID:    output.OrdemServicoID,
		ValorItemServicos: output.ValorItemServicos,
		ValorItemPecas:    output.ValorItemPecas,
		ValorTotal:        output.ValorTotal,
		Observacoes:       output.Observacoes,
		ItensServico:      itensServico,
		ItensPeca:         itensPeca,
	}
}

func toFluxoResponse(output app.FluxoOrcamentoOutput) FluxoOrcamentoResponse {
	return FluxoOrcamentoResponse{
		OrdemServicoID: output.OrdemServicoID,
		Status:         output.Status,
	}
}
