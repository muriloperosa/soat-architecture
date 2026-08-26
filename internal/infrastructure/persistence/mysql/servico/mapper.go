package servico

import (
	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// toModel converte a entidade de domínio pro model GORM.
func toModel(s *domainservico.Servico) *Model {
	return &Model{
		ID:                   s.ID(),
		Nome:                 s.Nome(),
		Descricao:            s.Descricao(),
		PrecoBase:            s.PrecoBase(),
		TempoEstimadoMinutos: s.TempoEstimado().Minutos(),
		CriadoPor:            s.CriadoPor(),
		Ativo:                s.Ativo(),
		DataCadastro:         s.DataCadastro(),
		DataAtualizacao:      s.DataAtualizacao(),
	}
}

// toEntity reidrata a entidade de domínio a partir do model (reconstituição,
// não revalida invariantes).
func toEntity(m *Model) *domainservico.Servico {
	return domainservico.RestaurarServico(
		m.ID,
		m.Nome,
		m.Descricao,
		m.PrecoBase,
		shared.RestaurarDuracaoEstimada(m.TempoEstimadoMinutos),
		m.CriadoPor,
		m.Ativo,
		m.DataCadastro,
		m.DataAtualizacao,
	)
}
