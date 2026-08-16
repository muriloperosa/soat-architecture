package auth

import (
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
)

// toModel converte a entidade de domínio pro model GORM. rt.ID zero (novo
// registro) deixa m.ID zerado pro banco gerar via autoincrement.
func toModel(rt *domainauth.RefreshToken) *Model {
	return &Model{
		ID:             rt.ID,
		UsuarioID:      rt.UsuarioID,
		Tipo:           string(rt.Tipo),
		Papel:          string(rt.Papel),
		TokenHash:      rt.TokenHash,
		AccessTokenJti: rt.AccessTokenJti,
		ExpiraEm:       rt.ExpiraEm,
		RevogadoEm:     rt.RevogadoEm,
	}
}

// toEntity reidrata a entidade de domínio a partir do model (reconstituição,
// não gera novo ID).
func toEntity(m *Model) *domainauth.RefreshToken {
	return &domainauth.RefreshToken{
		ID:             m.ID,
		UsuarioID:      m.UsuarioID,
		Tipo:           domainauth.TipoUsuario(m.Tipo),
		Papel:          domainauth.PapelUsuario(m.Papel),
		TokenHash:      m.TokenHash,
		AccessTokenJti: m.AccessTokenJti,
		ExpiraEm:       m.ExpiraEm,
		RevogadoEm:     m.RevogadoEm,
	}
}
