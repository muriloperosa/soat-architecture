package servico

// CriarServicoRequest é o corpo HTTP de POST /v1/servicos.
// PrecoBase é ponteiro pra binding:"required" aceitar 0.00 (preço gratuito).
type CriarServicoRequest struct {
	Nome                 string   `json:"nome" binding:"required" example:"Troca de óleo"`
	Descricao            string   `json:"descricao" binding:"required" example:"Troca de óleo e filtro"`
	PrecoBase            *float64 `json:"preco_base" binding:"required" example:"150.5"`
	TempoEstimadoMinutos int      `json:"tempo_estimado_minutos" binding:"required" example:"60"`
}

// AtualizarServicoRequest é o corpo HTTP de PUT /v1/servicos/:id.
type AtualizarServicoRequest struct {
	Nome                 string   `json:"nome" binding:"required" example:"Alinhamento"`
	Descricao            string   `json:"descricao" binding:"required" example:"Alinhamento e balanceamento"`
	PrecoBase            *float64 `json:"preco_base" binding:"required" example:"200.75"`
	TempoEstimadoMinutos int      `json:"tempo_estimado_minutos" binding:"required" example:"90"`
}

// ServicoResponse é a resposta comum de criação/atualização/consulta de serviço.
type ServicoResponse struct {
	ID                   uint64  `json:"id" example:"1"`
	Nome                 string  `json:"nome" example:"Troca de óleo"`
	Descricao            string  `json:"descricao" example:"Troca de óleo e filtro"`
	PrecoBase            float64 `json:"preco_base" example:"150.5"`
	TempoEstimadoMinutos int     `json:"tempo_estimado_minutos" example:"60"`
	CriadoPor            uint64  `json:"criado_por" example:"1"`
	Ativo                bool    `json:"ativo" example:"true"`
}
