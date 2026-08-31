package veiculo

import (
	"context"
	"errors"

	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
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

func (r *Repository) Salvar(ctx context.Context, veiculo *domain.Veiculo) error {
	model := toModel(veiculo)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	veiculo.AtribuirID(model.ID)

	return nil
}

func (r *Repository) BuscarPorID(ctx context.Context, id uint64) (*domain.Veiculo, error) {
	var model Model

	err := r.db.WithContext(ctx).First(&model, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVeiculoNaoEncontrado
		}

		return nil, err
	}

	return toDomain(model), nil
}

func (r *Repository) BuscarPorPlaca(ctx context.Context, placa domain.Placa) (*domain.Veiculo, error) {
	var model Model

	err := r.db.WithContext(ctx).Where("placa = ?", placa.String()).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVeiculoNaoEncontrado
		}

		return nil, err
	}

	return toDomain(model), nil
}

func (r *Repository) Listar(
	ctx context.Context,
	params domainquery.Params,
) (domainquery.Page[*domain.Veiculo], error) {
	normalized, err := r.queryBuilder.Normalize(params)
	if err != nil {
		return domainquery.Page[*domain.Veiculo]{}, err
	}

	filtered, err := r.queryBuilder.ApplyFilters(
		r.db.WithContext(ctx).Model(&Model{}),
		normalized.Filters,
	)
	if err != nil {
		return domainquery.Page[*domain.Veiculo]{}, err
	}

	var total int64

	if err = filtered.Count(&total).Error; err != nil {
		return domainquery.Page[*domain.Veiculo]{}, err
	}

	models := make([]Model, 0)

	if err = r.queryBuilder.
		ApplyPagination(filtered, normalized).
		Find(&models).
		Error; err != nil {
		return domainquery.Page[*domain.Veiculo]{}, err
	}

	veiculos := make([]*domain.Veiculo, 0, len(models))

	for _, model := range models {
		veiculos = append(veiculos, toDomain(model))
	}

	return domainquery.Page[*domain.Veiculo]{
		Items:      veiculos,
		Total:      total,
		Page:       normalized.Page,
		PageSize:   r.queryBuilder.PageSize(),
		TotalPages: r.queryBuilder.TotalPages(total),
		Order:      normalized.Order,
		Direction:  normalized.Direction,
	}, nil
}

func (r *Repository) Atualizar(ctx context.Context, veiculo *domain.Veiculo) error {
	model := toModel(veiculo)

	result := r.db.
		WithContext(ctx).
		Model(&Model{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"marca":               model.Marca,
			"modelo":              model.Modelo,
			"cor":                 model.Cor,
			"quilometragem_atual": model.QuilometragemAtual,
			"ativo":               model.Ativo,
			"data_atualizacao":    model.DataAtualizacao,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrVeiculoNaoEncontrado
	}

	return nil
}
