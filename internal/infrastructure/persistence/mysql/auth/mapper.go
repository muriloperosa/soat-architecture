package auth

import (
	"strconv"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
)

// toModel converte a entidade de domínio pro model GORM. rt.ID vazio (novo
// registro) deixa m.ID zerado pro banco gerar via autoincrement; rt.ID
// preenchido (atualização) faz o parse de volta pra uint64.
func toModel(rt *domainauth.RefreshToken) *Model {
	m := &Model{
		UsuarioID:      rt.UsuarioID,
		Tipo:           string(rt.Tipo),
		Papel:          string(rt.Papel),
		TokenHash:      rt.TokenHash,
		AccessTokenJti: rt.AccessTokenJti,
		ExpiraEm:       rt.ExpiraEm,
		RevogadoEm:     rt.RevogadoEm,
	}
	if rt.ID != "" {
		if id, err := strconv.ParseUint(rt.ID, 10, 64); err == nil {
			m.ID = id
		}
	}
	return m
}

// toEntity reidrata a entidade de domínio a partir do model (reconstituição,
// não gera novo ID).
func toEntity(m *Model) *domainauth.RefreshToken {
	return &domainauth.RefreshToken{
		ID:             strconv.FormatUint(m.ID, 10),
		UsuarioID:      m.UsuarioID,
		Tipo:           domainauth.TipoUsuario(m.Tipo),
		Papel:          domainauth.PapelUsuario(m.Papel),
		TokenHash:      m.TokenHash,
		AccessTokenJti: m.AccessTokenJti,
		ExpiraEm:       m.ExpiraEm,
		RevogadoEm:     m.RevogadoEm,
	}
}
