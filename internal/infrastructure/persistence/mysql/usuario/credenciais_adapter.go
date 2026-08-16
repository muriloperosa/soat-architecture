package usuario

import (
	"context"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// CredenciaisAdapter implementa domainauth.CredenciaisRepository sobre
// domainusuario.UsuarioRepository. Sem acesso a banco próprio, só mapeia
// Usuario pra Credencial. Existe porque as duas interfaces têm um
// BuscarPorEmail com retornos diferentes (não cabem no mesmo tipo Go) e
// domain/auth não pode depender do domínio usuario.
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
