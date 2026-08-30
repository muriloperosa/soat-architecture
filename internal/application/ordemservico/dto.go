package ordemservico

import (
	"time"

	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
)

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

type EntregarOrdemServicoInput struct {
	OrdemServicoID uint64
	UsuarioID      uint64
}

type HistoricoStatusOutput struct {
	ID          uint64
	Status      string
	AlteradoPor uint64
	Motivo      string
	AlteradoEm  time.Time
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
	CriadoPor            uint64
	DataCadastro         time.Time
	DataAtualizacao      time.Time
	HistoricoStatus      []HistoricoStatusOutput
}

type OrcamentoResumoOutput struct {
	ID                uint64
	ValorItemServicos float64
	ValorItemPecas    float64
	ValorTotal        float64
	Observacoes       string
}

type OrdemServicoResumoOutput struct {
	ID                   uint64
	Numero               string
	ClienteID            uint64
	VeiculoID            uint64
	QuilometragemEntrada uint32
	Status               string
	Diagnostico          string
	Observacoes          string
	CriadoPor            uint64
	DataCadastro         time.Time
	DataAtualizacao      time.Time
	Orcamento            *OrcamentoResumoOutput
}

type ListarOrdensServicoInput struct {
	appquery.ParamsInput
	SolicitanteID   uint64
	TipoSolicitante domainauth.TipoUsuario
}

type ListarOrdensServicoOutput struct {
	Items      []OrdemServicoResumoOutput
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
	Order      string
	Direction  string
}

func toOutput(os *domain.OrdemServico) OrdemServicoOutput {
	historicos := os.HistoricoStatus()
	historicoOutput := make([]HistoricoStatusOutput, 0, len(historicos))
	for _, historico := range historicos {
		historicoOutput = append(historicoOutput, HistoricoStatusOutput{
			ID:          historico.ID(),
			Status:      historico.Status().String(),
			AlteradoPor: historico.AlteradoPor(),
			Motivo:      historico.Motivo(),
			AlteradoEm:  historico.AlteradoEm(),
		})
	}

	return OrdemServicoOutput{
		ID:                   os.ID(),
		Numero:               os.Numero().String(),
		ClienteID:            os.ClienteID(),
		VeiculoID:            os.VeiculoID(),
		QuilometragemEntrada: uint32(os.QuilometragemEntrada()), // #nosec G115 -- domínio valida 0 <= km <= MaxUint32 na construção da OS
		Status:               os.Status().String(),
		Diagnostico:          os.Diagnostico(),
		Observacoes:          os.Observacoes(),
		CriadoPor:            os.CriadoPor(),
		DataCadastro:         os.DataCadastro(),
		DataAtualizacao:      os.DataAtualizacao(),
		HistoricoStatus:      historicoOutput,
	}
}

func toResumoOutput(os *domain.OrdemServico) OrdemServicoResumoOutput {
	return OrdemServicoResumoOutput{
		ID:                   os.ID(),
		Numero:               os.Numero().String(),
		ClienteID:            os.ClienteID(),
		VeiculoID:            os.VeiculoID(),
		QuilometragemEntrada: uint32(os.QuilometragemEntrada()), // #nosec G115 -- domínio valida 0 <= km <= MaxUint32 na construção da OS
		Status:               os.Status().String(),
		Diagnostico:          os.Diagnostico(),
		Observacoes:          os.Observacoes(),
		CriadoPor:            os.CriadoPor(),
		DataCadastro:         os.DataCadastro(),
		DataAtualizacao:      os.DataAtualizacao(),
	}
}

func toOrcamentoResumoOutput(o *domainorcamento.Orcamento) *OrcamentoResumoOutput {
	if o == nil {
		return nil
	}

	return &OrcamentoResumoOutput{
		ID:                o.ID(),
		ValorItemServicos: o.ValorItemServicos(),
		ValorItemPecas:    o.ValorItemPecas(),
		ValorTotal:        o.ValorTotal(),
		Observacoes:       o.Observacoes(),
	}
}
