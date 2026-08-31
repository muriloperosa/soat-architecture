package peca

import "time"

type Model struct {
	ID                  uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Codigo              string    `gorm:"column:codigo"`
	Nome                string    `gorm:"column:nome"`
	Marca               string    `gorm:"column:marca"`
	Descricao           string    `gorm:"column:descricao"`
	Preco               float64   `gorm:"column:preco"`
	QuantidadeEmEstoque int       `gorm:"column:quantidade_em_estoque"`
	EstoqueMinimo       int       `gorm:"column:estoque_minimo"`
	CriadoPor           uint64    `gorm:"column:criado_por"`
	Ativo               bool      `gorm:"column:ativo"`
	DataCadastro        time.Time `gorm:"column:data_cadastro"`
	DataAtualizacao     time.Time `gorm:"column:data_atualizacao"`
}

func (Model) TableName() string {
	return "pecas"
}
