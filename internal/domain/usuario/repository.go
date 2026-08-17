package usuario

import "context"

// UsuarioRepository persiste e consulta usuários internos.
// BuscarPorID e BuscarPorEmail retornam ErrUsuarioNaoEncontrado (sentinel,
// não gorm.ErrRecordNotFound cru) quando o usuário não existe.
type UsuarioRepository interface {
	Salvar(ctx context.Context, u *Usuario) error
	BuscarPorID(ctx context.Context, id uint64) (*Usuario, error)
	BuscarPorEmail(ctx context.Context, email string) (*Usuario, error)
	Atualizar(ctx context.Context, u *Usuario) error
}
