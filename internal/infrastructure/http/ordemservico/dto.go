package ordemservico

type AbrirOrdemServicoRequest struct {
	ClienteID            uint64 `json:"cliente_id" binding:"required" example:"1"`
	VeiculoID            uint64 `json:"veiculo_id" binding:"required" example:"1"`
	QuilometragemEntrada uint32 `json:"quilometragem_entrada" example:"52300"`
	Observacoes          string `json:"observacoes" example:"Cliente relatou ruído no motor"`
}

type OrdemServicoResponse struct {
	ID                   uint64 `json:"id" example:"1"`
	Numero               string `json:"numero" example:"OS-20260826-a1b2c3d4e5f6"`
	ClienteID            uint64 `json:"cliente_id" example:"1"`
	VeiculoID            uint64 `json:"veiculo_id" example:"1"`
	QuilometragemEntrada uint32 `json:"quilometragem_entrada" example:"52300"`
	Status               string `json:"status" example:"RECEBIDA"`
	Diagnostico          string `json:"diagnostico" example:""`
	Observacoes          string `json:"observacoes" example:"Cliente relatou ruído no motor"`
}
