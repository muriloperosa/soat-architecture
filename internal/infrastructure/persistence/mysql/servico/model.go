package servico

import "time"

// Model é a struct de persistência GORM do serviço do catálogo. Sem lógica de negócio.
type Model struct {
	ID                   uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Nome                 string    `gorm:"column:nome;size:150;not null"`
	Descricao            string    `gorm:"column:descricao;size:500;not null"`
	PrecoBase            float64   `gorm:"column:preco_base;type:decimal(10,2);not null"`
	TempoEstimadoMinutos int       `gorm:"column:tempo_estimado_minutos;not null"`
	CriadoPor            uint64    `gorm:"column:criado_por;not null"`
	Ativo                bool      `gorm:"column:ativo;not null"`
	DataCadastro         time.Time `gorm:"column:data_cadastro;not null"`
	DataAtualizacao      time.Time `gorm:"column:data_atualizacao;not null"`
}

func (Model) TableName() string {
	return "servicos"
}
