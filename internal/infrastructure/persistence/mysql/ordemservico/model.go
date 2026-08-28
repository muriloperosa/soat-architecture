package ordemservico

import "time"

type OrdemServicoModel struct {
	ID                   uint64                 `gorm:"column:id;primaryKey;autoIncrement"`
	Numero               string                 `gorm:"column:numero"`
	ClienteID            uint64                 `gorm:"column:cliente_id"`
	VeiculoID            uint64                 `gorm:"column:veiculo_id"`
	QuilometragemEntrada uint32                 `gorm:"column:quilometragem_entrada"`
	Status               string                 `gorm:"column:status"`
	Diagnostico          string                 `gorm:"column:diagnostico"`
	Observacoes          string                 `gorm:"column:observacoes"`
	CriadoPor            uint64                 `gorm:"column:criado_por"`
	DataCadastro         time.Time              `gorm:"column:data_cadastro"`
	DataAtualizacao      time.Time              `gorm:"column:data_atualizacao"`
	Historicos           []HistoricoStatusModel `gorm:"foreignKey:OrdemServicoID"`
}

func (OrdemServicoModel) TableName() string { return "ordens_servico" }

type HistoricoStatusModel struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OrdemServicoID uint64    `gorm:"column:ordem_servico_id"`
	Status         string    `gorm:"column:status"`
	AlteradoPor    uint64    `gorm:"column:alterado_por"`
	Motivo         string    `gorm:"column:motivo"`
	AlteradoEm     time.Time `gorm:"column:alterado_em"`
}

func (HistoricoStatusModel) TableName() string { return "historicos_status" }
