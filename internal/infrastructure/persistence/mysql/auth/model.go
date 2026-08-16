package auth

import "time"

// Model é a struct de persistência GORM do refresh token. Sem lógica de negócio.
type Model struct {
	ID             uint64     `gorm:"column:id;primaryKey;autoIncrement"`
	UsuarioID      uint64     `gorm:"column:usuario_id;not null;index:idx_refresh_tokens_usuario_id"`
	Tipo           string     `gorm:"column:tipo;type:enum('interno','cliente');not null"`
	Papel          string     `gorm:"column:papel;size:20;not null"`
	TokenHash      string     `gorm:"column:token_hash;size:64;not null;uniqueIndex:uk_refresh_tokens_token_hash"`
	AccessTokenJti string     `gorm:"column:access_token_jti;size:43;not null;index:idx_refresh_tokens_access_token_jti"`
	ExpiraEm       time.Time  `gorm:"column:expira_em;not null"`
	RevogadoEm     *time.Time `gorm:"column:revogado_em"`
}

func (Model) TableName() string {
	return "refresh_tokens"
}
