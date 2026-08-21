package servico

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domainservico "github.com/muriloperosa/soat-architecture/internal/domain/servico"
)

// Repository implementa domainservico.ServicoRepository via GORM/MySQL.
type Repository struct {
	db *gorm.DB
}

// NewServicoRepository monta o Repository sobre a conexão db informada.
func NewServicoRepository(db *gorm.DB) domainservico.ServicoRepository {
	return &Repository{db: db}
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
func (r *Repository) Listar(ctx context.Context) ([]*domainservico.Servico, error) {
	var models []Model
	if err := r.db.WithContext(ctx).Find(&models).Error; err != nil {
		return nil, err
	}
	servicos := make([]*domainservico.Servico, 0, len(models))
	for i := range models {
		servicos = append(servicos, toEntity(&models[i]))
	}
	return servicos, nil
}

// Atualizar persiste os campos mutáveis do serviço (nome, descricao, preco_base,
// tempo_estimado_minutos, ativo, data_atualizacao). Select("*") evita o GORM
// ignorar zero values (false, 0, "").
func (r *Repository) Atualizar(ctx context.Context, s *domainservico.Servico) error {
	m := toModel(s)
	return r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", m.ID).Select("*").Updates(m).Error
}
