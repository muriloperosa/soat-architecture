package auth

import (
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// RefreshToken é o registro persistido de um refresh token emitido.
// TokenHash nunca guarda o token bruto, só o hash (ver infrastructure/auth/jwt.go).
type RefreshToken struct {
	ID             uint64
	UsuarioID      uint64
	Tipo           TipoUsuario
	Papel          shared.PapelUsuario
	TokenHash      string
	AccessTokenJti string
	ExpiraEm       time.Time
	RevogadoEm     *time.Time
}

// EstaValido indica se o refresh token ainda pode ser usado: não revogado e não expirado.
func (r RefreshToken) EstaValido() bool {
	if r.RevogadoEm != nil {
		return false
	}
	return time.Now().Before(r.ExpiraEm)
}
