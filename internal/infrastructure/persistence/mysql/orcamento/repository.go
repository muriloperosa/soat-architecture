package orcamento

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

func NewOrcamentoRepository(db *gorm.DB) domain.OrcamentoRepository {
	return &Repository{db: db}
}

func (r *Repository) Salvar(ctx context.Context, o *domain.Orcamento) error {
	model := toModel(o)

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("ItensServico", "ItensPeca").Create(model).Error; err != nil {
			return err
		}

		if err := inserirNovosItens(tx, o, model.ID); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	o.AtribuirID(model.ID)
	return nil
}

func (r *Repository) BuscarPorOrdemServicoID(ctx context.Context, ordemServicoID uint64) (*domain.Orcamento, error) {
	var model OrcamentoModel
	if err := r.consultaComItens(ctx).Where("ordem_servico_id = ?", ordemServicoID).First(&model).Error; err != nil {
		return nil, traduzirErroConsulta(err)
	}

	return toDomain(model), nil
}

// BuscarPorOrdensServicoIDs busca os orçamentos associados às OS informadas em
// uma única consulta. Os itens não são carregados porque a listagem de OS usa
// somente o resumo financeiro do orçamento.
func (r *Repository) BuscarPorOrdensServicoIDs(ctx context.Context, ordensServicoIDs []uint64) ([]*domain.Orcamento, error) {
	if len(ordensServicoIDs) == 0 {
		return []*domain.Orcamento{}, nil
	}

	models := make([]OrcamentoModel, 0)
	if err := r.db.WithContext(ctx).
		Where("ordem_servico_id IN ?", ordensServicoIDs).
		Find(&models).Error; err != nil {
		return nil, err
	}

	orcamentos := make([]*domain.Orcamento, 0, len(models))
	for _, model := range models {
		orcamentos = append(orcamentos, toDomain(model))
	}

	return orcamentos, nil
}

// Atualizar persiste os dados escalares do orçamento e sincroniza os itens:
// insere os itens novos (ID == 0) e remove do banco os itens que não estão
// mais presentes no agregado (removidos via RemoverItemServico/RemoverItemPeca).
// Itens existentes preservam o ID, pois nunca são atualizados, só
// adicionados ou removidos.
func (r *Repository) Atualizar(ctx context.Context, o *domain.Orcamento) error {
	model := toModel(o)

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.
			Model(&OrcamentoModel{}).
			Where("id = ?", model.ID).
			Updates(map[string]any{
				"valor_item_servicos": model.ValorItemServicos,
				"valor_item_pecas":    model.ValorItemPecas,
				"valor_total":         model.ValorTotal,
				"observacoes":         model.Observacoes,
				"atualizado_em":       model.AtualizadoEm,
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domain.ErrOrcamentoNaoEncontrado
		}

		if err := sincronizarItensServico(tx, o, model.ID); err != nil {
			return err
		}
		if err := sincronizarItensPeca(tx, o, model.ID); err != nil {
			return err
		}

		return nil
	})
}

func inserirNovosItens(tx *gorm.DB, o *domain.Orcamento, orcamentoID uint64) error {
	itensServico := make([]ItemServicoModel, 0, len(o.ItensServico()))
	for _, item := range o.ItensServico() {
		itensServico = append(itensServico, toItemServicoModel(item, orcamentoID))
	}
	if len(itensServico) > 0 {
		if err := tx.Create(&itensServico).Error; err != nil {
			return err
		}
	}

	itensPeca := make([]ItemPecaModel, 0, len(o.ItensPeca()))
	for _, item := range o.ItensPeca() {
		itensPeca = append(itensPeca, toItemPecaModel(item, orcamentoID))
	}
	if len(itensPeca) > 0 {
		if err := tx.Create(&itensPeca).Error; err != nil {
			return err
		}
	}

	return nil
}

func sincronizarItensServico(tx *gorm.DB, o *domain.Orcamento, orcamentoID uint64) error {
	var idsAtuais []uint64
	if err := tx.Model(&ItemServicoModel{}).Where("orcamento_id = ?", orcamentoID).Pluck("id", &idsAtuais).Error; err != nil {
		return err
	}

	idsMantidos := make(map[uint64]bool)
	novos := make([]ItemServicoModel, 0)
	for _, item := range o.ItensServico() {
		if item.ID() == 0 {
			novos = append(novos, toItemServicoModel(item, orcamentoID))
		} else {
			idsMantidos[item.ID()] = true
		}
	}

	idsRemovidos := make([]uint64, 0)
	for _, id := range idsAtuais {
		if !idsMantidos[id] {
			idsRemovidos = append(idsRemovidos, id)
		}
	}

	if len(idsRemovidos) > 0 {
		if err := tx.Delete(&ItemServicoModel{}, idsRemovidos).Error; err != nil {
			return err
		}
	}
	if len(novos) > 0 {
		if err := tx.Create(&novos).Error; err != nil {
			return err
		}
	}

	return nil
}

func sincronizarItensPeca(tx *gorm.DB, o *domain.Orcamento, orcamentoID uint64) error {
	var idsAtuais []uint64
	if err := tx.Model(&ItemPecaModel{}).Where("orcamento_id = ?", orcamentoID).Pluck("id", &idsAtuais).Error; err != nil {
		return err
	}

	idsMantidos := make(map[uint64]bool)
	novos := make([]ItemPecaModel, 0)
	for _, item := range o.ItensPeca() {
		if item.ID() == 0 {
			novos = append(novos, toItemPecaModel(item, orcamentoID))
		} else {
			idsMantidos[item.ID()] = true
		}
	}

	idsRemovidos := make([]uint64, 0)
	for _, id := range idsAtuais {
		if !idsMantidos[id] {
			idsRemovidos = append(idsRemovidos, id)
		}
	}

	if len(idsRemovidos) > 0 {
		if err := tx.Delete(&ItemPecaModel{}, idsRemovidos).Error; err != nil {
			return err
		}
	}
	if len(novos) > 0 {
		if err := tx.Create(&novos).Error; err != nil {
			return err
		}
	}

	return nil
}

func (r *Repository) consultaComItens(ctx context.Context) *gorm.DB {
	return r.db.WithContext(ctx).
		Preload("ItensServico").
		Preload("ItensPeca")
}

func traduzirErroConsulta(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.ErrOrcamentoNaoEncontrado
	}
	return err
}
