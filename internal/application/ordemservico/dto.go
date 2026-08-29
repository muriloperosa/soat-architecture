package ordemservico

import domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"

type AbrirOrdemServicoInput struct {
	ClienteID            uint64
	VeiculoID            uint64
	QuilometragemEntrada uint32
	Observacoes          string
	UsuarioID            uint64
}

type IniciarDiagnosticoInput struct {
	OrdemServicoID uint64
	UsuarioID      uint64
}

type InformarDiagnosticoInput struct {
	OrdemServicoID uint64
	Diagnostico    string
}

type IniciarExecucaoInput struct {
	OrdemServicoID uint64
	UsuarioID      uint64
}

type OrdemServicoOutput struct {
	ID                   uint64
	Numero               string
	ClienteID            uint64
	VeiculoID            uint64
	QuilometragemEntrada uint32
	Status               string
	Diagnostico          string
	Observacoes          string
}

func toOutput(os *domain.OrdemServico) OrdemServicoOutput {
	return OrdemServicoOutput{
		ID:                   os.ID(),
		Numero:               os.Numero().String(),
		ClienteID:            os.ClienteID(),
		VeiculoID:            os.VeiculoID(),
		QuilometragemEntrada: uint32(os.QuilometragemEntrada()), // #nosec G115 -- domínio valida 0 <= km <= MaxUint32 na construção da OS
		Status:               os.Status().String(),
		Diagnostico:          os.Diagnostico(),
		Observacoes:          os.Observacoes(),
	}
}
