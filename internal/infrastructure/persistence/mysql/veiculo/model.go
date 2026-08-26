package veiculo

import "time"

type VeiculoModel struct {
	ID                 uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	Placa              string    `gorm:"column:placa"`
	Marca              string    `gorm:"column:marca"`
	Modelo             string    `gorm:"column:modelo"`
	QuilometragemAtual uint32    `gorm:"column:quilometragem_atual"`
	Ano                uint16    `gorm:"column:ano"`
	Cor                string    `gorm:"column:cor"`
	CriadoPor          uint64    `gorm:"column:criado_por"`
	Ativo              bool      `gorm:"column:ativo"`
	DataCadastro       time.Time `gorm:"column:data_cadastro"`
	DataAtualizacao    time.Time `gorm:"column:data_atualizacao"`
}

func (VeiculoModel) TableName() string {
	return "veiculos"
}
