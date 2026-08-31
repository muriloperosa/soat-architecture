package peca

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"

	"gorm.io/gorm"
)

type Repository struct {
	db           *gorm.DB
	queryBuilder *mysqlquery.Builder
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(db *gorm.DB) domain.Repository {
	builder := NewQueryBuilder()

	return &Repository{
		db:           db,
		queryBuilder: builder,
	}
}

// Salvar implements [peca.Repository].
func (r *Repository) Salvar(ctx context.Context, peca *domain.Peca) error {
	model := toModel(peca)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	peca.AtribuirID(model.ID)

	return nil
}

// BuscarPorID implements [peca.Repository].
func (r *Repository) BuscarPorID(ctx context.Context, id uint64) (*domain.Peca, error) {
	var model PecaModel

	err := r.db.WithContext(ctx).First(&model, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPecaNaoEncontrada
		}

		return nil, err
	}

	return toDomain(model), nil
}

// BuscarPorCodigo implements [peca.Repository].
func (r *Repository) BuscarPorCodigo(ctx context.Context, codigo string) (*domain.Peca, error) {
	var model PecaModel

	err := r.db.WithContext(ctx).Where("codigo = ?", codigo).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPecaNaoEncontrada
		}

		return nil, err
	}

	return toDomain(model), nil
}

// BuscarPorIDComBloqueio implements [peca.Repository].
func (r *Repository) BuscarPorIDComBloqueio(ctx context.Context, id uint64) (*domain.Peca, error) {
	var model PecaModel

	db := mysql.ComBloqueio(mysql.DBFromContext(ctx, r.db))

	err := db.WithContext(ctx).First(&model, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPecaNaoEncontrada
		}

		return nil, err
	}

	return toDomain(model), nil
}

// Listar implements [peca.Repository].
func (r *Repository) Listar(
	ctx context.Context,
	params domainquery.Params,
) (domainquery.Page[*domain.Peca], error) {
	normalized, err := r.queryBuilder.Normalize(params)
	if err != nil {
		return domainquery.Page[*domain.Peca]{}, err
	}

	filtered, err := r.queryBuilder.ApplyFilters(
		r.db.WithContext(ctx).Model(&PecaModel{}),
		normalized.Filters,
	)
	if err != nil {
		return domainquery.Page[*domain.Peca]{}, err
	}

	var total int64

	if err = filtered.Count(&total).Error; err != nil {
		return domainquery.Page[*domain.Peca]{}, err
	}

	models := make([]PecaModel, 0)

	if err = r.queryBuilder.
		ApplyPagination(filtered, normalized).
		Find(&models).
		Error; err != nil {
		return domainquery.Page[*domain.Peca]{}, err
	}

	pecas := make([]*domain.Peca, 0, len(models))

	for _, model := range models {
		pecas = append(pecas, toDomain(model))
	}

	return domainquery.Page[*domain.Peca]{
		Items:      pecas,
		Total:      total,
		Page:       normalized.Page,
		PageSize:   r.queryBuilder.PageSize(),
		TotalPages: r.queryBuilder.TotalPages(total),
		Order:      normalized.Order,
		Direction:  normalized.Direction,
	}, nil
}

// Atualizar implements [peca.Repository].
func (r *Repository) Atualizar(ctx context.Context, peca *domain.Peca) error {
	model := toModel(peca)

	result := mysql.DBFromContext(ctx, r.db).
		WithContext(ctx).
		Model(&PecaModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"nome":                  model.Nome,
			"marca":                 model.Marca,
			"descricao":             model.Descricao,
			"preco":                 model.Preco,
			"quantidade_em_estoque": model.QuantidadeEmEstoque,
			"estoque_minimo":        model.EstoqueMinimo,
			"ativo":                 model.Ativo,
			"data_atualizacao":      model.DataAtualizacao,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrPecaNaoEncontrada
	}

	return nil
}
