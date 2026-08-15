package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
)

type Repository struct {
	db *gorm.DB
}

func NewRepositorioRefreshToken(db *gorm.DB) domainauth.RepositorioRefreshToken {
	return &Repository{db: db}
}

func (r *Repository) Salvar(ctx context.Context, rt *domainauth.RefreshToken) error {
	m := toModel(rt)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	rt.ID = m.ID
	return nil
}

func (r *Repository) BuscarPorHash(ctx context.Context, tokenHash string) (*domainauth.RefreshToken, error) {
	var m Model
	err := r.db.WithContext(ctx).Where("token_hash = ?", tokenHash).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if err != nil {
		return nil, err
	}
	return toEntity(&m), nil
}

func (r *Repository) Revogar(ctx context.Context, id string) error {
	now := gorm.Expr("NOW()")
	return r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Update("revogado_em", now).Error
}
