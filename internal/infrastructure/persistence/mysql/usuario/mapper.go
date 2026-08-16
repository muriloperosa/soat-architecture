package usuario

import (
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// toModel converte a entidade de domínio pro model GORM.
func toModel(u *domainusuario.Usuario) *Model {
	return &Model{
		ID:                 u.ID(),
		Papel:              string(u.Papel()),
		Nome:               u.Nome(),
		Email:              u.Email().String(),
		SenhaHash:          u.Senha().String(),
		RequerAlterarSenha: u.RequerAlterarSenha(),
		Ativo:              u.Ativo(),
		DataCadastro:       u.DataCadastro(),
		DataAtualizacao:    u.DataAtualizacao(),
	}
}

// toEntity reidrata a entidade de domínio a partir do model (reconstituição,
// não revalida nem gera novo hash de senha).
func toEntity(m *Model) (*domainusuario.Usuario, error) {
	email, err := shared.NewEmail(m.Email)
	if err != nil {
		return nil, err
	}
	senha := shared.RestaurarSenhaHash(m.SenhaHash)
	u := domainusuario.RestaurarUsuario(m.ID, m.Nome, email, senha, shared.PapelUsuario(m.Papel), m.RequerAlterarSenha, m.Ativo, m.DataCadastro, m.DataAtualizacao)
	return u, nil
}
