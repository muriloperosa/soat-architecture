package peca

// CadastrarPecaRequest é o corpo HTTP de POST /v1/pecas.
type CadastrarPecaRequest struct {
	Nome                string  `json:"nome" binding:"required" example:"Pastilha de freio"`
	Marca               string  `json:"marca" binding:"required" example:"Bosch"`
	Descricao           string  `json:"descricao" binding:"required" example:"Pastilha de freio dianteira"`
	Preco               float64 `json:"preco" binding:"required" example:"89.9"`
	QuantidadeEmEstoque int     `json:"quantidade_em_estoque" example:"20"`
	EstoqueMinimo       int     `json:"estoque_minimo" example:"5"`
}

// AtualizarPecaRequest é o corpo HTTP de PUT /v1/pecas/:id.
type AtualizarPecaRequest struct {
	Nome          string  `json:"nome" binding:"required" example:"Pastilha de freio"`
	Marca         string  `json:"marca" binding:"required" example:"Bosch"`
	Descricao     string  `json:"descricao" binding:"required" example:"Pastilha de freio dianteira"`
	Preco         float64 `json:"preco" binding:"required" example:"89.9"`
	EstoqueMinimo int     `json:"estoque_minimo" example:"5"`
}

// ReporEstoqueRequest é o corpo HTTP de PATCH /v1/pecas/:id/repor-estoque.
type ReporEstoqueRequest struct {
	Quantidade int `json:"quantidade" binding:"required" example:"10"`
}

// PecaResponse é a resposta comum de criação/atualização/consulta de peça.
type PecaResponse struct {
	ID                  uint64  `json:"id" example:"1"`
	Codigo              string  `json:"codigo" example:"PC1J3K2L1M0N"`
	Nome                string  `json:"nome" example:"Pastilha de freio"`
	Marca               string  `json:"marca" example:"Bosch"`
	Descricao           string  `json:"descricao" example:"Pastilha de freio dianteira"`
	Preco               float64 `json:"preco" example:"89.9"`
	QuantidadeEmEstoque int     `json:"quantidade_em_estoque" example:"20"`
	EstoqueMinimo       int     `json:"estoque_minimo" example:"5"`
	CriadoPor           uint64  `json:"criado_por" example:"1"`
	Ativo               bool    `json:"ativo" example:"true"`
}

type ListarPecasResponse struct {
	Items      []PecaResponse `json:"items"`
	Total      int64          `json:"total" example:"42"`
	Page       int            `json:"page" example:"1"`
	PageSize   int            `json:"page_size" example:"20"`
	TotalPages int            `json:"total_pages" example:"3"`
	Order      string         `json:"order" example:"nome"`
	Direction  string         `json:"direction" example:"ASC"`
}
