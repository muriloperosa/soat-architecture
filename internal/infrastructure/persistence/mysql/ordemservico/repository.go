package ordemservico

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainquery "github.com/muriloperosa/soat-architecture/internal/domain/query"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
	mysqlquery "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/query"
	"gorm.io/gorm"
)

type Repository struct {
	db           *gorm.DB
	queryBuilder *mysqlquery.Builder
}

func NewOrdemServicoRepository(db *gorm.DB) domain.OrdemServicoRepository {
	return &Repository{
		db:           db,
		queryBuilder: NewQueryBuilder(),
	}
}

func (r *Repository) Salvar(ctx context.Context, os *domain.OrdemServico) error {
	model := toModel(os)
	var ids []uint64

	err := mysql.DBFromContext(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		ids = make([]uint64, len(models))
		for i, m := range models {
			ids[i] = m.ID
		}

		return nil
	})
	if err != nil {
		return err
	}

	os.AtribuirID(model.ID)
	os.AtribuirIDsHistoricoPendente(ids)
	return nil
}

func (r *Repository) BuscarPorID(ctx context.Context, id uint64) (*domain.OrdemServico, error) {
	var model OrdemServicoModel
	if err := r.consultaComHistorico(ctx).First(&model, id).Error; err != nil {
		return nil, traduzirErroConsulta(err)
	}

	return toDomain(model)
}

func (r *Repository) BuscarPorIDComBloqueio(ctx context.Context, id uint64) (*domain.OrdemServico, error) {
	var model OrdemServicoModel
	if err := mysql.ComBloqueio(r.consultaComHistorico(ctx)).First(&model, id).Error; err != nil {
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

func (r *Repository) Listar(
	ctx context.Context,
	params domainquery.Params,
) (domainquery.Page[*domain.OrdemServico], error) {
	normalized, err := r.queryBuilder.Normalize(params)
	if err != nil {
		return domainquery.Page[*domain.OrdemServico]{}, err
	}

	filtered, err := r.queryBuilder.ApplyFilters(
		mysql.DBFromContext(ctx, r.db).WithContext(ctx).Model(&OrdemServicoModel{}),
		normalized.Filters,
	)
	if err != nil {
		return domainquery.Page[*domain.OrdemServico]{}, err
	}

	var total int64
	if err = filtered.Count(&total).Error; err != nil {
		return domainquery.Page[*domain.OrdemServico]{}, err
	}

	models := make([]OrdemServicoModel, 0)
	if err = r.queryBuilder.ApplyPagination(filtered, normalized).Find(&models).Error; err != nil {
		return domainquery.Page[*domain.OrdemServico]{}, err
	}

	ordens := make([]*domain.OrdemServico, 0, len(models))
	for _, model := range models {
		ordem, mapperErr := toDomain(model)
		if mapperErr != nil {
			return domainquery.Page[*domain.OrdemServico]{}, mapperErr
		}
		ordens = append(ordens, ordem)
	}

	return domainquery.Page[*domain.OrdemServico]{
		Items:      ordens,
		Total:      total,
		Page:       normalized.Page,
		PageSize:   r.queryBuilder.PageSize(),
		TotalPages: r.queryBuilder.TotalPages(total),
		Order:      normalized.Order,
		Direction:  normalized.Direction,
	}, nil
}

func (r *Repository) Atualizar(ctx context.Context, os *domain.OrdemServico) error {
	model := toModel(os)
	var ids []uint64

	err := mysql.DBFromContext(ctx, r.db).WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

		ids = make([]uint64, len(novosHistoricos))
		for i, m := range novosHistoricos {
			ids[i] = m.ID
		}

		return nil
	})
	if err != nil {
		return err
	}

	os.AtribuirIDsHistoricoPendente(ids)
	return nil
}

func (r *Repository) consultaComHistorico(ctx context.Context) *gorm.DB {
	return mysql.DBFromContext(ctx, r.db).WithContext(ctx).Preload("Historicos", func(db *gorm.DB) *gorm.DB {
		return db.Order("alterado_em ASC")
	})
}

func traduzirErroConsulta(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrOrdemServicoNaoEncontrada
	}
	return err
}
