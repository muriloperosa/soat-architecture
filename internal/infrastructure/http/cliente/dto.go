package cliente

// CriarClienteRequest é o corpo HTTP de POST /v1/clientes.
type CriarClienteRequest struct {
	Nome       string `json:"nome" binding:"required" example:"Maria Silva"`
	Email      string `json:"email" binding:"required" example:"maria@oficina.com"`
	Senha      string `json:"senha" binding:"required" example:"senha123"`
	Documento  string `json:"documento" binding:"required" example:"12345678909"`
	TipoPessoa string `json:"tipo_pessoa" binding:"required" example:"PF" enums:"PF,PJ"`
	Telefone   string `json:"telefone" binding:"required" example:"11999998888"`
}

// AtualizarClienteRequest é o corpo HTTP de PUT /v1/clientes/:id.
type AtualizarClienteRequest struct {
	Nome      string `json:"nome" binding:"required" example:"Maria Silva"`
	Email     string `json:"email" binding:"required" example:"maria@oficina.com"`
	SenhaNova string `json:"senha_nova,omitempty" example:"novaSenha123"`
	Telefone  string `json:"telefone" binding:"required" example:"11999998888"`
}

type AlterarSenhaRequest struct {
	SenhaNova string `json:"senha_nova" binding:"required"`
}

// ClienteResponse é a resposta comum de criação/atualização/consulta de cliente.
type ClienteResponse struct {
	ID                 uint64 `json:"id" example:"1"`
	Nome               string `json:"nome" example:"Maria Silva"`
	Email              string `json:"email" example:"maria@oficina.com"`
	Documento          string `json:"documento" example:"12345678909"`
	TipoPessoa         string `json:"tipo_pessoa" example:"PF" enums:"PF,PJ"`
	Telefone           string `json:"telefone" example:"11999998888"`
	Ativo              bool   `json:"ativo" example:"true"`
	RequerAlterarSenha bool   `json:"requer_alterar_senha" example:"true"`
	CriadoPor          uint64 `json:"criado_por" example:"1"`
}

// ListarClientesResponse contém os clientes e os metadados da página.
type ListarClientesResponse struct {
	Items     []ClienteResponse `json:"items"`
	Total     int64             `json:"total" example:"42"`
	Offset    int               `json:"offset" example:"0"`
	Limit     int               `json:"limit" example:"20"`
	Order     string            `json:"order" example:"nome"`
	Direction string            `json:"direction" example:"ASC"`
}
