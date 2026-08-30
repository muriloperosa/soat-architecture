package servico

import (
	"context"
	"errors"

	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"gorm.io/gorm"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"
)

// Repository implementa domainservico.ServicoRepository via GORM/MySQL.
type Repository struct {
	db           *gorm.DB
	queryBuilder *mysqlquery.Builder
}

// NewServicoRepository monta o Repository sobre a conexão db informada.
func NewServicoRepository(db *gorm.DB) domainservico.ServicoRepository {
	builder := NewQueryBuilder()

	return &Repository{
		db:           db,
		queryBuilder: builder,
	}
}

// Salvar insere o serviço e preenche s.ID com o ID gerado pelo banco (autoincrement).
func (r *Repository) Salvar(ctx context.Context, s *domainservico.Servico) error {
	m := toModel(s)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	s.AtribuirID(m.ID)
	return nil
}

// BuscarPorID busca o serviço pelo ID. Devolve domainservico.ErrServicoNaoEncontrado
// (sentinel de domínio, não gorm.ErrRecordNotFound cru) quando não existe.
func (r *Repository) BuscarPorID(ctx context.Context, id uint64) (*domainservico.Servico, error) {
	var m Model
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainservico.ErrServicoNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return toEntity(&m), nil
}

// Listar devolve todos os serviços do catálogo.
func (r *Repository) Listar(
	ctx context.Context,
	params domainquery.Params,
) (domainquery.Page[*domainservico.Servico], error) {
	normalized, err := r.queryBuilder.Normalize(params)
	if err != nil {
		return domainquery.Page[*domainservico.Servico]{}, err
	}

	filtered, err := r.queryBuilder.ApplyFilters(
		r.db.WithContext(ctx).Model(&Model{}),
		normalized.Filters,
	)
	if err != nil {
		return domainquery.Page[*domainservico.Servico]{}, err
	}

	var total int64

	if err = filtered.Count(&total).Error; err != nil {
		return domainquery.Page[*domainservico.Servico]{}, err
	}

	models := make([]Model, 0)

	if err = r.queryBuilder.
		ApplyPagination(filtered, normalized).
		Find(&models).
		Error; err != nil {
		return domainquery.Page[*domainservico.Servico]{}, err
	}

	servicos := make([]*domainservico.Servico, 0, len(models))

	for i := range models {
		servicos = append(servicos, toEntity(&models[i]))
	}

	return domainquery.Page[*domainservico.Servico]{
		Items:      servicos,
		Total:      total,
		Page:       normalized.Page,
		PageSize:   r.queryBuilder.PageSize(),
		TotalPages: r.queryBuilder.TotalPages(total),
		Order:      normalized.Order,
		Direction:  normalized.Direction,
	}, nil
}

// Atualizar persiste os campos mutáveis do serviço (nome, descricao, preco_base,
// tempo_estimado_minutos, ativo, data_atualizacao).
func (r *Repository) Atualizar(ctx context.Context, s *domainservico.Servico) error {
	m := toModel(s)

	result := r.db.
		WithContext(ctx).
		Model(&Model{}).
		Where("id = ?", m.ID).
		Updates(map[string]any{
			"nome":                   m.Nome,
			"descricao":              m.Descricao,
			"preco_base":             m.PrecoBase,
			"tempo_estimado_minutos": m.TempoEstimadoMinutos,
			"ativo":                  m.Ativo,
			"data_atualizacao":       m.DataAtualizacao,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domainservico.ErrServicoNaoEncontrado
	}

	return nil
}
