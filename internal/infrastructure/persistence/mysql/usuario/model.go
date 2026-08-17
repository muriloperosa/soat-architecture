package usuario

import "time"

// Model é a struct de persistência GORM do usuário interno. Sem lógica de negócio.
type Model struct {
	ID                 uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Papel              string    `gorm:"column:papel;size:20;not null"`
	Nome               string    `gorm:"column:nome;size:150;not null"`
	Email              string    `gorm:"column:email;size:255;not null;uniqueIndex:uk_usuarios_email"`
	SenhaHash          string    `gorm:"column:senha_hash;size:255;not null"`
	RequerAlterarSenha bool      `gorm:"column:requer_alterar_senha;not null"`
	Ativo              bool      `gorm:"column:ativo;not null"`
	DataCadastro       time.Time `gorm:"column:data_cadastro;not null"`
	DataAtualizacao    time.Time `gorm:"column:data_atualizacao;not null"`
}

func (Model) TableName() string {
	return "usuarios"
}
