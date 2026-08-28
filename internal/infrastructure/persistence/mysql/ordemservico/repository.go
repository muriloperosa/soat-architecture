package ordemservico

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

var _ domain.OrdemServicoRepository = (*Repository)(nil)

func NewOrdemServicoRepository(db *gorm.DB) domain.OrdemServicoRepository {
	return &Repository{db: db}
}

func (r *Repository) Salvar(ctx context.Context, os *domain.OrdemServico) error {
	model := toModel(os)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Historicos").Create(model).Error; err != nil {
			return err
		}

		historicos := os.HistoricoStatus()
		models := make([]HistoricoStatusModel, 0, len(historicos))
		for _, historico := range historicos {
			models = append(models, toHistoricoModel(historico, model.ID))
		}

		if len(models) > 0 {
			if err := tx.Create(&models).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	os.AtribuirID(model.ID)
	return nil
}

func (r *Repository) BuscarPorID(ctx context.Context, id uint64) (*domain.OrdemServico, error) {
	var model OrdemServicoModel
	if err := r.consultaComHistorico(ctx).First(&model, id).Error; err != nil {
		return nil, traduzirErroConsulta(err)
	}

	return toDomain(model)
}

func (r *Repository) BuscarPorNumero(ctx context.Context, numero string) (*domain.OrdemServico, error) {
	var model OrdemServicoModel
	if err := r.consultaComHistorico(ctx).Where("numero = ?", numero).First(&model).Error; err != nil {
		return nil, traduzirErroConsulta(err)
	}

	return toDomain(model)
}

func (r *Repository) Atualizar(ctx context.Context, os *domain.OrdemServico) error {
	model := toModel(os)
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&OrdemServicoModel{}).
			Where("id = ?", model.ID).
			Updates(map[string]any{
				"status":           model.Status,
				"diagnostico":      model.Diagnostico,
				"observacoes":      model.Observacoes,
				"data_atualizacao": model.DataAtualizacao,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrOrdemServicoNaoEncontrada
		}

		novosHistoricos := make([]HistoricoStatusModel, 0)
		for _, historico := range os.HistoricoStatus() {
			if historico.ID() == 0 {
				novosHistoricos = append(novosHistoricos, toHistoricoModel(historico, model.ID))
			}
		}

		if len(novosHistoricos) > 0 {
			if err := tx.Create(&novosHistoricos).Error; err != nil {
				return err
			}
		}

		return nil
	})
}

func (r *Repository) consultaComHistorico(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).Preload("Historicos", func(db *gorm.DB) *gorm.DB {
		return db.Order("alterado_em ASC")
	})
}

func traduzirErroConsulta(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrOrdemServicoNaoEncontrada
	}
	return err
}
