package ordemservico

import app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"

func toInput(usuarioID uint64, request AbrirOrdemServicoRequest) app.AbrirOrdemServicoInput {
	return app.AbrirOrdemServicoInput{
		ClienteID:            request.ClienteID,
		VeiculoID:            request.VeiculoID,
		QuilometragemEntrada: request.QuilometragemEntrada,
		Observacoes:          request.Observacoes,
		UsuarioID:            usuarioID,
	}
}

func toResponse(output app.OrdemServicoOutput) OrdemServicoResponse {
	return OrdemServicoResponse{
		ID:                   output.ID,
		Numero:               output.Numero,
		ClienteID:            output.ClienteID,
		VeiculoID:            output.VeiculoID,
		QuilometragemEntrada: output.QuilometragemEntrada,
		Status:               output.Status,
		Diagnostico:          output.Diagnostico,
		Observacoes:          output.Observacoes,
	}
}
