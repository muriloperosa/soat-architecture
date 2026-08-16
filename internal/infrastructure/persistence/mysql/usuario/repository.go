package usuario

import (
	"context"
	"errors"

	"gorm.io/gorm"

	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// Repository implementa domainusuario.UsuarioRepository via GORM/MySQL.
type Repository struct {
	db *gorm.DB
}

// NewUsuarioRepository monta o Repository sobre a conexão db informada.
func NewUsuarioRepository(db *gorm.DB) domainusuario.UsuarioRepository {
	return &Repository{db: db}
}

// Salvar insere o usuário e preenche u.ID com o ID gerado pelo banco (autoincrement).
func (r *Repository) Salvar(ctx context.Context, u *domainusuario.Usuario) error {
	m := toModel(u)
	if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
		return err
	}
	u.AtribuirID(m.ID)
	return nil
}

// BuscarPorID busca o usuário pelo ID. Devolve domainusuario.ErrUsuarioNaoEncontrado
// (sentinel de domínio, não gorm.ErrRecordNotFound cru) quando não existe.
func (r *Repository) BuscarPorID(ctx context.Context, id uint64) (*domainusuario.Usuario, error) {
	var m Model
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainusuario.ErrUsuarioNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return toEntity(&m)
}

// BuscarPorEmail busca o usuário pelo email. Mesmo contrato de erro de BuscarPorID.
func (r *Repository) BuscarPorEmail(ctx context.Context, email string) (*domainusuario.Usuario, error) {
	var m Model
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&m).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domainusuario.ErrUsuarioNaoEncontrado
	}
	if err != nil {
		return nil, err
	}
	return toEntity(&m)
}

// Atualizar persiste todos os campos mutáveis do usuário (nome, papel, senha,
// requer_alterar_senha, ativo, data_atualizacao).
func (r *Repository) Atualizar(ctx context.Context, u *domainusuario.Usuario) error {
	m := toModel(u)
	return r.db.WithContext(ctx).Model(&Model{}).Where("id = ?", m.ID).Updates(m).Error
}
