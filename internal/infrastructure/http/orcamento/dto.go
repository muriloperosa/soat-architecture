package orcamento

type GerarOrcamentoRequest struct {
	Observacoes string `json:"observacoes" example:"Aguardando aprovação do cliente"`
}

type AdicionarServicoOrcamentoRequest struct {
	ServicoID  uint64 `json:"servico_id" binding:"required" example:"1"`
	Quantidade int    `json:"quantidade" binding:"required" example:"1"`
}

type AdicionarPecaOrcamentoRequest struct {
	PecaID     uint64 `json:"peca_id" binding:"required" example:"1"`
	Quantidade int    `json:"quantidade" binding:"required" example:"2"`
}

type AlterarQuantidadePecaOrcamentoRequest struct {
	Quantidade int `json:"quantidade" binding:"required" example:"3"`
}

type ItemServicoResponse struct {
	ID                   uint64  `json:"id" example:"1"`
	ServicoID            uint64  `json:"servico_id" example:"1"`
	Quantidade           int     `json:"quantidade" example:"1"`
	Valor                float64 `json:"valor" example:"150.00"`
	TempoEstimadoMinutos int     `json:"tempo_estimado_minutos" example:"60"`
	Subtotal             float64 `json:"subtotal" example:"150.00"`
}

type ItemPecaResponse struct {
	ID         uint64  `json:"id" example:"1"`
	PecaID     uint64  `json:"peca_id" example:"1"`
	Descricao  string  `json:"descricao" example:"Filtro de óleo"`
	Quantidade int     `json:"quantidade" example:"2"`
	Valor      float64 `json:"valor" example:"50.00"`
	Subtotal   float64 `json:"subtotal" example:"100.00"`
}

type OrcamentoResponse struct {
	ID                uint64                `json:"id" example:"1"`
	OrdemServicoID    uint64                `json:"ordem_servico_id" example:"1"`
	ValorItemServicos float64               `json:"valor_item_servicos" example:"150.00"`
	ValorItemPecas    float64               `json:"valor_item_pecas" example:"100.00"`
	ValorTotal        float64               `json:"valor_total" example:"250.00"`
	Observacoes       string                `json:"observacoes" example:"Aguardando aprovação do cliente"`
	ItensServico      []ItemServicoResponse `json:"itens_servico"`
	ItensPeca         []ItemPecaResponse    `json:"itens_peca"`
}

type RejeitarOrcamentoRequest struct {
	Motivo string `json:"motivo" binding:"required" example:"Valor acima do esperado"`
}

type FluxoOrcamentoResponse struct {
	OrdemServicoID uint64 `json:"ordem_servico_id" example:"1"`
	Status         string `json:"status" example:"APROVADA"`
}
