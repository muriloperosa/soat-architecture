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

func toIniciarDiagnosticoInput(ordemServicoID, usuarioID uint64) app.IniciarDiagnosticoInput {
	return app.IniciarDiagnosticoInput{
		OrdemServicoID: ordemServicoID,
		UsuarioID:      usuarioID,
	}
}

func toInformarDiagnosticoInput(
	ordemServicoID uint64,
	request InformarDiagnosticoRequest,
) app.InformarDiagnosticoInput {
	return app.InformarDiagnosticoInput{
		OrdemServicoID: ordemServicoID,
		Diagnostico:    request.Diagnostico,
	}
}

func toIniciarExecucaoInput(ordemServicoID, usuarioID uint64) app.IniciarExecucaoInput {
	return app.IniciarExecucaoInput{
		OrdemServicoID: ordemServicoID,
		UsuarioID:      usuarioID,
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
