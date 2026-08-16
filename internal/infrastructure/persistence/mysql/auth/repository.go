package auth

import (
	"context"
	"errors"
	"strconv"

	"gorm.io/gorm"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
)

// Repository implementa domainauth.RefreshTokenRepository via GORM/MySQL.
type Repository struct {
	db *gorm.DB
}

// NewRefreshTokenRepository monta o Repository sobre a conexão db informada.
func NewRefreshTokenRepository(db *gorm.DB) domainauth.RefreshTokenRepository {
	return &Repository{db: db}
}

// Salvar insere o refresh token e preenche rt.ID com o ID gerado pelo banco
// (autoincrement).
func (r *Repository) Salvar(ctx context.Context, rt *domainauth.RefreshToken) error {
	m := toModel(rt)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	rt.ID = strconv.FormatUint(m.ID, 10)
	return nil
}

// BuscarPorHash busca o refresh token pelo hash do token bruto. Devolve
// gorm.ErrRecordNotFound (sem encapsular) quando não existe.
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

// Revogar marca o refresh token de id como revogado (revogado_em = agora).
func (r *Repository) Revogar(ctx context.Context, id string) error {
	now := gorm.Expr("NOW()")
	return r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", id).Update("revogado_em", now).Error
}

// AccessTokenRevogado indica se existe um refresh token revogado cujo
// access_token_jti seja jti, usado por AuthenticationMiddleware pra
// rejeitar access tokens ainda não expirados mas já revogados em par.
func (r *Repository) AccessTokenRevogado(ctx context.Context, jti string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&Model{}).
		Where("access_token_jti = ? AND revogado_em IS NOT NULL", jti).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}
