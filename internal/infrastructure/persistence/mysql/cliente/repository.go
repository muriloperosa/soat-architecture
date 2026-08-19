package cliente

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"

	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

var _ domain.ClienteRepository = (*Repository)(nil)

func NewClienteRepository(db *gorm.DB) domain.ClienteRepository {
	return &Repository{db: db}
}

// NewRepository mantém compatibilidade. Prefira NewClienteRepository.
func NewRepository(db *gorm.DB) domain.ClienteRepository {
	return NewClienteRepository(db)
}

// Salvar implements [cliente.Repository].
func (r *Repository) Salvar(ctx context.Context, cliente *domain.Cliente) error {
	model := toModel(*cliente)

	if err := r.db.WithContext(ctx).Create(&model).Error; err != nil {
		return err
	}

	cliente.DefinirID(model.ID)

	return nil
}

// BuscarPorID implements [cliente.Repository].
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

// BuscarPorDocumento implements [cliente.Repository].
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

// Atualizar implements [cliente.Repository].
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
