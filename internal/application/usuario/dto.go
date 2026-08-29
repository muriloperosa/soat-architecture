package usuario

import (
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// CriarUsuarioInput é o DTO de entrada do CriarUsuarioUseCase.
type CriarUsuarioInput struct {
	Nome         string
	Email        string
	SenhaInicial string
	Papel        string
}

// AtualizarUsuarioInput é o DTO de entrada do AtualizarUsuarioUseCase.
// SenhaNova é opcional: vazio significa "não trocar senha"; quando
// informado, é o admin redefinindo a senha (força troca no próximo login).
type AtualizarUsuarioInput struct {
	ID        uint64
	Nome      string
	Email     string
	SenhaNova string
	Papel     string
}

// AlterarSenhaInput é o DTO de entrada do AlterarSenhaUseCase.
type AlterarSenhaInput struct {
	UsuarioID uint64
	SenhaNova string
}

// UsuarioOutput é o DTO de saída comum aos use cases de gestão/consulta de usuário.
type UsuarioOutput struct {
	ID                 uint64
	Nome               string
	Email              string
	Papel              shared.PapelUsuario
	Ativo              bool
	RequerAlterarSenha bool
}

func toOutput(u *domainusuario.Usuario) UsuarioOutput {
	return UsuarioOutput{
		ID:                 u.ID(),
		Nome:               u.Nome(),
		Email:              u.Email().String(),
		Papel:              u.Papel(),
		Ativo:              u.Ativo(),
		RequerAlterarSenha: u.RequerAlterarSenha(),
	}
}
