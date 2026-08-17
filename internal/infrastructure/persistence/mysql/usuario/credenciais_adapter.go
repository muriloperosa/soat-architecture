package usuario

import (
	"context"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// CredenciaisAdapter implementa domainauth.CredenciaisRepository e
// domainauth.UsuarioStatusRepository sobre domainusuario.UsuarioRepository.
// Sem acesso a banco próprio, só mapeia Usuario pras projeções que domain/auth
// entende. Existe porque as interfaces têm BuscarPorEmail/EstaAtivo com
// retornos diferentes dos de UsuarioRepository (não cabem no mesmo tipo Go)
// e domain/auth não pode depender do domínio usuario.
type CredenciaisAdapter struct {
	usuarios domainusuario.UsuarioRepository
}

// NewCredenciaisAdapter monta o adapter sobre o UsuarioRepository informado.
func NewCredenciaisAdapter(usuarios domainusuario.UsuarioRepository) *CredenciaisAdapter {
	return &CredenciaisAdapter{usuarios: usuarios}
}

// BuscarPorEmail busca o usuário por email e mapeia pra Credencial. Mesmo
// contrato de erro de UsuarioRepository.BuscarPorEmail (ErrUsuarioNaoEncontrado
// sentinel quando não existe).
func (a *CredenciaisAdapter) BuscarPorEmail(ctx context.Context, email string) (*domainauth.Credencial, error) {
	u, err := a.usuarios.BuscarPorEmail(ctx, email)
	if err != nil {
		return nil, err
	}

	return toCredencial(*u), nil
}

// EstaAtivo indica se o usuário de id id ainda está ativo. Mesmo contrato de
// erro de UsuarioRepository.BuscarPorID (ErrUsuarioNaoEncontrado sentinel
// quando não existe — tratado como inativo por AuthenticationMiddleware).
func (a *CredenciaisAdapter) EstaAtivo(ctx context.Context, id uint64) (bool, error) {
	u, err := a.usuarios.BuscarPorID(ctx, id)
	if err != nil {
		return false, err
	}
	return u.Ativo(), nil
}
