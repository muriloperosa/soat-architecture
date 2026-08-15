package auth

import (
	"github.com/google/uuid"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
)

func toModel(rt *domainauth.RefreshToken) *Model {
	id := rt.ID
	if id == "" {
		id = uuid.NewString()
	}
	return &Model{
		ID:         id,
		UsuarioID:  rt.UsuarioID,
		Tipo:       string(rt.Tipo),
		Papel:      string(rt.Papel),
		TokenHash:  rt.TokenHash,
		ExpiraEm:   rt.ExpiraEm,
		RevogadoEm: rt.RevogadoEm,
	}
}

// toEntity reidrata a entidade de domínio a partir do model (reconstituição,
// não gera novo ID).
func toEntity(m *Model) *domainauth.RefreshToken {
	return &domainauth.RefreshToken{
		ID:         m.ID,
		UsuarioID:  m.UsuarioID,
		Tipo:       domainauth.TipoUsuario(m.Tipo),
		Papel:      domainauth.PapelUsuario(m.Papel),
		TokenHash:  m.TokenHash,
		ExpiraEm:   m.ExpiraEm,
		RevogadoEm: m.RevogadoEm,
	}
}
