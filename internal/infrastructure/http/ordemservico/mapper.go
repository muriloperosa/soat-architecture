package ordemservico

import (
	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httpquery"
)

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
	return app.IniciarDiagnosticoInput{OrdemServicoID: ordemServicoID, UsuarioID: usuarioID}
}

func toInformarDiagnosticoInput(ordemServicoID uint64, request InformarDiagnosticoRequest) app.InformarDiagnosticoInput {
	return app.InformarDiagnosticoInput{OrdemServicoID: ordemServicoID, Diagnostico: request.Diagnostico}
}

func toIniciarExecucaoInput(ordemServicoID, usuarioID uint64) app.IniciarExecucaoInput {
	return app.IniciarExecucaoInput{OrdemServicoID: ordemServicoID, UsuarioID: usuarioID}
}

func toFinalizarInput(ordemServicoID, usuarioID uint64) app.FinalizarOrdemServicoInput {
	return app.FinalizarOrdemServicoInput{OrdemServicoID: ordemServicoID, UsuarioID: usuarioID}
}

func toEntregarInput(ordemServicoID, usuarioID uint64) app.EntregarOrdemServicoInput {
	return app.EntregarOrdemServicoInput{OrdemServicoID: ordemServicoID, UsuarioID: usuarioID}
}

func toListarInput(params httpquery.Params, solicitanteID uint64, tipoSolicitante domainauth.TipoUsuario) app.ListarOrdensServicoInput {
	var filters []appquery.FilterInput
	if len(params.Filters) > 0 {
		filters = make([]appquery.FilterInput, 0, len(params.Filters))
		for _, filter := range params.Filters {
			filters = append(filters, appquery.FilterInput{Field: filter.Field, Operator: filter.Operator, Value: filter.Value})
		}
	}

	return app.ListarOrdensServicoInput{
		ParamsInput: appquery.ParamsInput{
			Page: params.Page, Order: params.Order, Direction: params.Direction, Filters: filters,
		},
		SolicitanteID:   solicitanteID,
		TipoSolicitante: tipoSolicitante,
	}
}

func toResponse(output app.OrdemServicoOutput) OrdemServicoResponse {
	historicos := make([]HistoricoStatusResponse, 0, len(output.HistoricoStatus))
	for _, historico := range output.HistoricoStatus {
		historicos = append(historicos, HistoricoStatusResponse{
			ID: historico.ID, Status: historico.Status, AlteradoPor: historico.AlteradoPor,
			Motivo: historico.Motivo, AlteradoEm: historico.AlteradoEm,
		})
	}

	return OrdemServicoResponse{
		ID: output.ID, Numero: output.Numero, ClienteID: output.ClienteID, VeiculoID: output.VeiculoID,
		QuilometragemEntrada: output.QuilometragemEntrada, Status: output.Status, Diagnostico: output.Diagnostico,
		Observacoes: output.Observacoes, CriadoPor: output.CriadoPor, DataCadastro: output.DataCadastro,
		DataAtualizacao: output.DataAtualizacao, HistoricoStatus: historicos,
	}
}

func toResumoResponse(output app.OrdemServicoResumoOutput) OrdemServicoResumoResponse {
	return OrdemServicoResumoResponse{
		ID: output.ID, Numero: output.Numero, ClienteID: output.ClienteID, VeiculoID: output.VeiculoID,
		QuilometragemEntrada: output.QuilometragemEntrada, Status: output.Status, Diagnostico: output.Diagnostico,
		Observacoes: output.Observacoes, CriadoPor: output.CriadoPor, DataCadastro: output.DataCadastro,
		DataAtualizacao: output.DataAtualizacao, Orcamento: toOrcamentoResumoResponse(output.Orcamento),
	}
}

func toOrcamentoResumoResponse(output *app.OrcamentoResumoOutput) *OrcamentoResumoResponse {
	if output == nil {
		return nil
	}

	return &OrcamentoResumoResponse{
		ID: output.ID, ValorItemServicos: output.ValorItemServicos, ValorItemPecas: output.ValorItemPecas,
		ValorTotal: output.ValorTotal, Observacoes: output.Observacoes,
	}
}

func toListResponse(output app.ListarOrdensServicoOutput) ListarOrdensServicoResponse {
	items := make([]OrdemServicoResumoResponse, 0, len(output.Items))
	for _, item := range output.Items {
		items = append(items, toResumoResponse(item))
	}

	return ListarOrdensServicoResponse{
		Items: items, Total: output.Total, Page: output.Page, PageSize: output.PageSize,
		TotalPages: output.TotalPages, Order: output.Order, Direction: output.Direction,
	}
}
