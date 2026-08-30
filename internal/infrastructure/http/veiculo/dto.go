package veiculo

type CadastrarVeiculoRequest struct {
	Placa              string `json:"placa" binding:"required" example:"ABC1D23"`
	Marca              string `json:"marca" binding:"required" example:"Fiat"`
	Modelo             string `json:"modelo" binding:"required" example:"Uno"`
	QuilometragemAtual uint32 `json:"quilometragem_atual" example:"15000"`
	Ano                uint16 `json:"ano" binding:"required" example:"2020"`
	Cor                string `json:"cor" binding:"required" example:"Prata"`
}

type AtualizarVeiculoRequest struct {
	Marca              string `json:"marca" binding:"required" example:"Fiat"`
	Modelo             string `json:"modelo" binding:"required" example:"Uno"`
	Cor                string `json:"cor" binding:"required" example:"Prata"`
	QuilometragemAtual uint32 `json:"quilometragem_atual" example:"15500"`
}

type VeiculoResponse struct {
	ID                 uint64 `json:"id" example:"1"`
	Placa              string `json:"placa" example:"ABC1D23"`
	Marca              string `json:"marca" example:"Fiat"`
	Modelo             string `json:"modelo" example:"Uno"`
	QuilometragemAtual uint32 `json:"quilometragem_atual" example:"15000"`
	Ano                uint16 `json:"ano" example:"2020"`
	Cor                string `json:"cor" example:"Prata"`
	CriadoPor          uint64 `json:"criado_por" example:"1"`
	Ativo              bool   `json:"ativo" example:"true"`
}

type ListarVeiculosResponse struct {
	Items      []VeiculoResponse `json:"items"`
	Total      int64             `json:"total" example:"42"`
	Page       int               `json:"page" example:"1"`
	PageSize   int               `json:"page_size" example:"20"`
	TotalPages int               `json:"total_pages" example:"3"`
	Order      string            `json:"order" example:"id"`
	Direction  string            `json:"direction" example:"ASC"`
}
