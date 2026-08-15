package auth

import "time"

// RefreshToken é o registro persistido de um refresh token emitido.
// TokenHash nunca guarda o token bruto — só o hash (ver infrastructure/auth/refresh_hash.go).
type RefreshToken struct {
	ID         string
	UsuarioID  string
	Tipo       TipoUsuario
	Papel      PapelUsuario
	TokenHash  string
	ExpiraEm   time.Time
	RevogadoEm *time.Time
}

// EstaValido indica se o refresh token ainda pode ser usado: não revogado e não expirado.
func (r RefreshToken) EstaValido() bool {
	if r.RevogadoEm != nil {
		return false
	}
	return time.Now().Before(r.ExpiraEm)
}
