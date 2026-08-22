package peca

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Repository struct {
	db *gorm.DB
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(db *gorm.DB) domain.Repository {
	return &Repository{db: db}
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

	db := mysql.DBFromContext(ctx, r.db)

	err := db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&model, id).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrPecaNaoEncontrada
		}

		return nil, err
	}

	return toDomain(model), nil
}

// Atualizar implements [peca.Repository].
func (r *Repository) Atualizar(ctx context.Context, peca *domain.Peca) error {
	model := toModel(peca)

	result := r.db.
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
