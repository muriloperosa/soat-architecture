package reservapeca

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

var _ domain.Repository = (*Repository)(nil)

func NewRepository(db *gorm.DB) domain.Repository {
	return &Repository{db: db}
}

// Salvar implements [reservapeca.Repository].
func (r *Repository) Salvar(ctx context.Context, reserva *domain.ReservaPeca) error {
	model := toModel(reserva)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	reserva.AtribuirID(model.ID)

	return nil
}

// Atualizar implements [reservapeca.Repository].
func (r *Repository) Atualizar(ctx context.Context, reserva *domain.ReservaPeca) error {
	model := toModel(reserva)

	result := r.db.
		WithContext(ctx).
		Model(&ReservaPecaModel{}).
		Where("id = ?", model.ID).
		Updates(map[string]any{
			"quantidade":    model.Quantidade,
			"atualizada_em": model.AtualizadaEm,
		})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrReservaNaoEncontrada
	}

	return nil
}

// BuscarPorOrdemEPeca implements [reservapeca.Repository].
func (r *Repository) BuscarPorOrdemEPeca(ctx context.Context, ordemServicoID, pecaID uint64) (*domain.ReservaPeca, error) {
	var model ReservaPecaModel

	err := r.db.WithContext(ctx).
		Where("ordem_servico_id = ? AND peca_id = ?", ordemServicoID, pecaID).
		First(&model).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, domain.ErrReservaNaoEncontrada
		}

		return nil, err
	}

	return toDomain(model), nil
}

// BuscarPorOrdemServico implements [reservapeca.Repository].
func (r *Repository) BuscarPorOrdemServico(ctx context.Context, ordemServicoID uint64) ([]*domain.ReservaPeca, error) {
	var models []ReservaPecaModel

	if err := r.db.WithContext(ctx).Where("ordem_servico_id = ?", ordemServicoID).Find(&models).Error; err != nil {
		return nil, err
	}

	reservas := make([]*domain.ReservaPeca, 0, len(models))
	for _, model := range models {
		reservas = append(reservas, toDomain(model))
	}

	return reservas, nil
}

// SomarQuantidadeReservada implements [reservapeca.Repository]. Soma a
// quantidade reservada de uma peça em todas as Ordens de Serviço.
func (r *Repository) SomarQuantidadeReservada(ctx context.Context, pecaID uint64) (int, error) {
	var total int

	err := r.db.WithContext(ctx).
		Model(&ReservaPecaModel{}).
		Where("peca_id = ?", pecaID).
		Select("COALESCE(SUM(quantidade), 0)").
		Scan(&total).Error

	if err != nil {
		return 0, err
	}

	return total, nil
}

// Remover implements [reservapeca.Repository].
func (r *Repository) Remover(ctx context.Context, ordemServicoID, pecaID uint64) error {
	result := r.db.WithContext(ctx).
		Where("ordem_servico_id = ? AND peca_id = ?", ordemServicoID, pecaID).
		Delete(&ReservaPecaModel{})

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return domain.ErrReservaNaoEncontrada
	}

	return nil
}
