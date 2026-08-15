package auth

import "context"

// RepositorioRefreshToken persiste e consulta refresh tokens.
// Implementado em internal/infrastructure/persistence/mysql/auth.
type RepositorioRefreshToken interface {
	Salvar(ctx context.Context, rt *RefreshToken) error
	BuscarPorHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revogar(ctx context.Context, id string) error
}
