package veiculo

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(db *gorm.DB) domain.Repository {
	return &Repository{db: db}
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
	var model VeiculoModel

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
	var model VeiculoModel

	err := r.db.WithContext(ctx).Where("placa = ?", placa.String()).First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrVeiculoNaoEncontrado
		}

		return nil, err
	}

	return toDomain(model), nil
}

func (r *Repository) Atualizar(ctx context.Context, veiculo *domain.Veiculo) error {
	model := toModel(veiculo)

	result := r.db.
		WithContext(ctx).
		Model(&VeiculoModel{}).
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
