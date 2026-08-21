package reservapeca

import "time"

type ReservaPecaModel struct {
	ID             uint64    `gorm:"column:id;primaryKey;autoIncrement"`
	OrdemServicoID uint64    `gorm:"column:ordem_servico_id"`
	PecaID         uint64    `gorm:"column:peca_id"`
	Quantidade     int       `gorm:"column:quantidade"`
	CriadaEm       time.Time `gorm:"column:criada_em"`
	AtualizadaEm   time.Time `gorm:"column:atualizada_em"`
}

func (ReservaPecaModel) TableName() string {
	return "reservas_pecas"
}
