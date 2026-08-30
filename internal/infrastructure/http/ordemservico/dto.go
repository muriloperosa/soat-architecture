package ordemservico

import "time"

type AbrirOrdemServicoRequest struct {
	ClienteID            uint64 `json:"cliente_id" binding:"required" example:"1"`
	VeiculoID            uint64 `json:"veiculo_id" binding:"required" example:"1"`
	QuilometragemEntrada uint32 `json:"quilometragem_entrada" example:"52300"`
	Observacoes          string `json:"observacoes" example:"Cliente relatou ruído no motor"`
}

type InformarDiagnosticoRequest struct {
	Diagnostico string `json:"diagnostico" example:"Falha na bomba de combustível"`
}

type HistoricoStatusResponse struct {
	ID          uint64    `json:"id" example:"1"`
	Status      string    `json:"status" example:"RECEBIDA"`
	AlteradoPor uint64    `json:"alterado_por" example:"1"`
	Motivo      string    `json:"motivo" example:""`
	AlteradoEm  time.Time `json:"alterado_em"`
}

type OrdemServicoResponse struct {
	ID                   uint64                    `json:"id" example:"1"`
	Numero               string                    `json:"numero" example:"OS-20260826-a1b2c3d4e5f6"`
	ClienteID            uint64                    `json:"cliente_id" example:"1"`
	VeiculoID            uint64                    `json:"veiculo_id" example:"1"`
	QuilometragemEntrada uint32                    `json:"quilometragem_entrada" example:"52300"`
	Status               string                    `json:"status" example:"RECEBIDA"`
	Diagnostico          string                    `json:"diagnostico" example:""`
	Observacoes          string                    `json:"observacoes" example:"Cliente relatou ruído no motor"`
	CriadoPor            uint64                    `json:"criado_por" example:"1"`
	DataCadastro         time.Time                 `json:"data_cadastro"`
	DataAtualizacao      time.Time                 `json:"data_atualizacao"`
	HistoricoStatus      []HistoricoStatusResponse `json:"historico_status"`
}

type OrdemServicoResumoResponse struct {
	ID                   uint64    `json:"id" example:"1"`
	Numero               string    `json:"numero" example:"OS-20260826-a1b2c3d4e5f6"`
	ClienteID            uint64    `json:"cliente_id" example:"1"`
	VeiculoID            uint64    `json:"veiculo_id" example:"1"`
	QuilometragemEntrada uint32    `json:"quilometragem_entrada" example:"52300"`
	Status               string    `json:"status" example:"RECEBIDA"`
	Diagnostico          string    `json:"diagnostico" example:""`
	Observacoes          string    `json:"observacoes" example:"Cliente relatou ruído no motor"`
	CriadoPor            uint64    `json:"criado_por" example:"1"`
	DataCadastro         time.Time `json:"data_cadastro"`
	DataAtualizacao      time.Time `json:"data_atualizacao"`
}

type ListarOrdensServicoResponse struct {
	Items      []OrdemServicoResumoResponse `json:"items"`
	Total      int64                        `json:"total" example:"42"`
	Page       int                          `json:"page" example:"1"`
	PageSize   int                          `json:"page_size" example:"20"`
	TotalPages int                          `json:"total_pages" example:"3"`
	Order      string                       `json:"order" example:"data_cadastro"`
	Direction  string                       `json:"direction" example:"DESC"`
}
