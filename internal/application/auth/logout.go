package auth

import (
	"context"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
)

// LogoutUseCase revoga um refresh token. Idempotente: se o token não existir
// ou já estiver revogado, não é erro (no-op).
type LogoutUseCase struct {
	refreshTokens domainauth.RepositorioRefreshToken
}

func NewLogoutUseCase(refreshTokens domainauth.RepositorioRefreshToken) *LogoutUseCase {
	return &LogoutUseCase{refreshTokens: refreshTokens}
}

func (uc *LogoutUseCase) Executar(ctx context.Context, input LogoutInput) error {
	hash := infraauth.HashRefreshToken(input.RefreshTokenBruto)
	rt, err := uc.refreshTokens.BuscarPorHash(ctx, hash)
	if err != nil {
		return nil
	}
	if rt.RevogadoEm != nil {
		return nil
	}
	return uc.refreshTokens.Revogar(ctx, rt.ID)
}
