package auth

import (
	"crypto/sha256"
	"encoding/hex"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// HashRefreshToken calcula o hash (SHA-256) do token bruto pra persistência e busca.
func HashRefreshToken(bruto string) string {
	sum := sha256.Sum256([]byte(bruto))
	return hex.EncodeToString(sum[:])
}

// RefreshToken é o registro persistido de um refresh token emitido.
// TokenHash nunca guarda o token bruto, só o hash (ver HashRefreshToken).
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
