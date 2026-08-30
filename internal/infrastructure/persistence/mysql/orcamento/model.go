package orcamento

import "time"

// OrcamentoModel é a struct de persistência GORM do orçamento. Sem lógica de negócio.
type OrcamentoModel struct {
	ID                uint64             `gorm:"column:id;primaryKey;autoIncrement"`
	OrdemServicoID    uint64             `gorm:"column:ordem_servico_id"`
	ValorItemServicos float64            `gorm:"column:valor_item_servicos"`
	ValorItemPecas    float64            `gorm:"column:valor_item_pecas"`
	ValorTotal        float64            `gorm:"column:valor_total"`
	Observacoes       string             `gorm:"column:observacoes"`
	CriadoPor         uint64             `gorm:"column:criado_por"`
	CriadoEm          time.Time          `gorm:"column:criado_em"`
	AtualizadoEm      time.Time          `gorm:"column:atualizado_em"`
	ItensServico      []ItemServicoModel `gorm:"foreignKey:OrcamentoID"`
	ItensPeca         []ItemPecaModel    `gorm:"foreignKey:OrcamentoID"`
}

func (OrcamentoModel) TableName() string { return "orcamentos" }

type ItemServicoModel struct {
	ID                   uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	OrcamentoID          uint64  `gorm:"column:orcamento_id"`
	ServicoID            uint64  `gorm:"column:servico_id"`
	Quantidade           int     `gorm:"column:quantidade"`
	Valor                float64 `gorm:"column:valor"`
	TempoEstimadoMinutos int     `gorm:"column:tempo_estimado_minutos"`
}

func (ItemServicoModel) TableName() string { return "orcamentos_servicos" }

type ItemPecaModel struct {
	ID          uint64  `gorm:"column:id;primaryKey;autoIncrement"`
	OrcamentoID uint64  `gorm:"column:orcamento_id"`
	PecaID      uint64  `gorm:"column:peca_id"`
	Descricao   string  `gorm:"column:descricao"`
	Quantidade  int     `gorm:"column:quantidade"`
	Valor       float64 `gorm:"column:valor"`
}

func (ItemPecaModel) TableName() string { return "orcamentos_pecas" }
