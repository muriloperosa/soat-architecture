package cliente

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/query"
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"

	"gorm.io/gorm"
)

type Repository struct {
	db           *gorm.DB
	queryBuilder *mysqlquery.Builder
}

var _ domain.ClienteRepository = (*Repository)(nil)

func NewClienteRepository(db *gorm.DB) domain.ClienteRepository {
	builder := NewQueryBuilder()
	return &Repository{db: db, queryBuilder: builder}
}

// NewRepository mantém compatibilidade. Prefira NewClienteRepository.
func NewRepository(db *gorm.DB) domain.ClienteRepository {
	return NewClienteRepository(db)
}

// Salvar implements [cliente.ClienteRepository].
func (r *Repository) Salvar(ctx context.Context, cliente *domain.Cliente) error {
	model := toModel(*cliente)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	cliente.DefinirID(model.ID)

	return nil
}

// BuscarPorID implements [cliente.ClienteRepository].
func (r *Repository) BuscarPorID(ctx context.Context, id uint64) (*domain.Cliente, error) {
	var model Model

	err := r.db.WithContext(ctx).First(&model, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrClienteNaoEncontrado
		}

		return nil, err
	}

	return toEntity(model)
}

// BuscarPorDocumento implements [cliente.ClienteRepository].
func (r *Repository) BuscarPorDocumento(ctx context.Context, documento string) (*domain.Cliente, error) {
	var model Model

	err := r.db.WithContext(ctx).Where("documento = ?", documento).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrClienteNaoEncontrado
		}

		return nil, err
	}

	return toEntity(model)
}

func (r *Repository) BuscarPorEmail(ctx context.Context, email string) (*domain.Cliente, error) {
	var model Model
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&model).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrClienteNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return toEntity(model)
}

// Listar implements [cliente.ClienteRepository].
func (r *Repository) Listar(
	ctx context.Context,
	params query.Params,
) (query.Page[*domain.Cliente], error) {
	normalized, err := r.queryBuilder.Normalize(params)
	if err != nil {
		return query.Page[*domain.Cliente]{}, err
	}

	filtered, err := r.queryBuilder.ApplyFilters(
		r.db.WithContext(ctx).Model(&Model{}),
		normalized.Filters,
	)
	if err != nil {
		return query.Page[*domain.Cliente]{}, err
	}

	var total int64
	if err = filtered.Count(&total).Error; err != nil {
		return query.Page[*domain.Cliente]{}, err
	}

	models := make([]Model, 0)
	if err = r.queryBuilder.ApplyPagination(filtered, normalized).Find(&models).Error; err != nil {
		return query.Page[*domain.Cliente]{}, err
	}

	clientes := make([]*domain.Cliente, 0, len(models))
	for _, model := range models {
		cliente, mapperErr := toEntity(model)
		if mapperErr != nil {
			return query.Page[*domain.Cliente]{}, mapperErr
		}
		clientes = append(clientes, cliente)
	}

	return query.Page[*domain.Cliente]{
		Items:     clientes,
		Total:     total,
		Offset:    normalized.Offset,
		Limit:     normalized.Limit,
		Order:     normalized.Order,
		Direction: normalized.Direction,
	}, nil
}

// Atualizar implements [cliente.ClienteRepository].
func (r *Repository) Atualizar(ctx context.Context, cliente *domain.Cliente) error {
	model := toModel(*cliente)

	result := r.db.
		WithContext(ctx).
		Model(&Model{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"documento":            model.Documento,
			"tipo":                 model.Tipo,
			"nome":                 model.Nome,
			"email":                model.Email,
			"senha_hash":           model.Senha,
			"requer_alterar_senha": model.RequerAlterarSenha,
			"telefone":             model.Telefone,
			"ativo":                model.Ativo,
			"data_atualizacao":     model.DataAtualizacao,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrClienteNaoEncontrado
	}

	return nil
}
