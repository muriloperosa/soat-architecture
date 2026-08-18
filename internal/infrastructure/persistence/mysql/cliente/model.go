package cliente

import "time"

type ClienteModel struct {
	ID                 uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Documento          string    `gorm:"column:documento"`
	Tipo               string    `gorm:"column:tipo"`
	Nome               string    `gorm:"column:nome"`
	Email              string    `gorm:"column:email"`
	Telefone           string    `gorm:"column:telefone"`
	Senha              string    `gorm:"column:senha_hash"`
	RequerAlterarSenha bool      `gorm:"column:requer_alterar_senha"`
	CriadoPor          uint64    `gorm:"column:criado_por"`
	Ativo              bool      `gorm:"column:ativo"`
	DataCadastro       time.Time `gorm:"column:data_cadastro"`
	DataAtualizacao    time.Time `gorm:"column:data_atualizacao"`
}

func (ClienteModel) TableName() string {
	return "clientes"
}
