package usuario

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

// CriarUsuarioRequest é o corpo HTTP de POST /v1/usuarios.
type CriarUsuarioRequest struct {
	Nome  string              `json:"nome" binding:"required" example:"Ana Souza"`
	Email string              `json:"email" binding:"required,email" example:"ana@oficina.com"`
	Senha string              `json:"senha" binding:"required" example:"senha123"`
	Papel shared.PapelUsuario `json:"papel" binding:"required" example:"MECANICO"`
}

// AtualizarUsuarioRequest é o corpo HTTP de PUT /v1/usuarios/:id.
// SenhaNova é opcional — quando informado, o admin redefine a senha do
// usuário (força troca no próximo login).
type AtualizarUsuarioRequest struct {
	Nome      string              `json:"nome" binding:"required" example:"Ana Souza"`
	Email     string              `json:"email" binding:"required,email" example:"ana@oficina.com"`
	SenhaNova string              `json:"senha_nova,omitempty" example:"novaSenha123"`
	Papel     shared.PapelUsuario `json:"papel" binding:"required" example:"ATENDENTE"`
}

// AlterarSenhaRequest é o corpo HTTP de PUT /v1/usuarios/me/senha.
type AlterarSenhaRequest struct {
	SenhaNova string `json:"senha_nova" example:"novaSenha123"`
}

// UsuarioResponse é a resposta comum de criação/atualização/consulta de usuário.
type UsuarioResponse struct {
	ID                 uint64              `json:"id" example:"1"`
	Nome               string              `json:"nome" example:"Ana Souza"`
	Email              string              `json:"email" example:"ana@oficina.com"`
	Papel              shared.PapelUsuario `json:"papel" example:"MECANICO"`
	Ativo              bool                `json:"ativo" example:"true"`
	RequerAlterarSenha bool                `json:"requer_alterar_senha" example:"true"`
}
