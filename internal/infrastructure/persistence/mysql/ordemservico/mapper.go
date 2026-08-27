package ordemservico

import domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"

func toModel(os *domain.OrdemServico) *OrdemServicoModel {
	return &OrdemServicoModel{
		ID:                   os.ID(),
		Numero:               os.Numero().String(),
		ClienteID:            os.ClienteID(),
		VeiculoID:            os.VeiculoID(),
		QuilometragemEntrada: uint32(os.QuilometragemEntrada()),
		Status:               os.Status().String(),
		Diagnostico:          os.Diagnostico(),
		Observacoes:          os.Observacoes(),
		CriadoPor:            os.CriadoPor(),
		DataCadastro:         os.DataCadastro(),
		DataAtualizacao:      os.DataAtualizacao(),
	}
}

func toHistoricoModel(h domain.HistoricoStatus, ordemServicoID uint64) HistoricoStatusModel {
	return HistoricoStatusModel{
		ID:             h.ID(),
		OrdemServicoID: ordemServicoID,
		Status:         h.Status().String(),
		AlteradoPor:    h.AlteradoPor(),
		Motivo:         h.Motivo(),
		AlteradoEm:     h.AlteradoEm(),
	}
}

func toDomain(model OrdemServicoModel) (*domain.OrdemServico, error) {
	numero, err := domain.NewNumeroOrdemServico(model.Numero)
	if err != nil {
		return nil, err
	}

	status, err := domain.NewStatusOrdemServico(model.Status)
	if err != nil {
		return nil, err
	}

	historicos := make([]domain.HistoricoStatus, 0, len(model.Historicos))
	for _, historicoModel := range model.Historicos {
		historicoStatus, err := domain.NewStatusOrdemServico(historicoModel.Status)
		if err != nil {
			return nil, err
		}

		historicos = append(historicos, domain.ReidratarHistoricoStatus(
			historicoModel.ID,
			historicoModel.OrdemServicoID,
			historicoStatus,
			historicoModel.AlteradoEm,
			historicoModel.AlteradoPor,
			historicoModel.Motivo,
		))
	}

	return domain.ReidratarOrdemServico(
		model.ID,
		numero,
		model.ClienteID,
		model.VeiculoID,
		int(model.QuilometragemEntrada),
		status,
		model.Diagnostico,
		model.Observacoes,
		model.CriadoPor,
		historicos,
		model.DataCadastro,
		model.DataAtualizacao,
	), nil
}
