package auth

import "time"

// Model é a struct de persistência GORM do refresh token. Sem lógica de negócio.
type Model struct {
	ID             uint64 `gorm:"primaryKey;autoIncrement"`
	UsuarioID      string
	Tipo           string
	Papel          string
	TokenHash      string `gorm:"uniqueIndex"`
	AccessTokenJti string `gorm:"index"`
	ExpiraEm       time.Time
	RevogadoEm     *time.Time
}

func (Model) TableName() string {
	return "refresh_tokens"
}
