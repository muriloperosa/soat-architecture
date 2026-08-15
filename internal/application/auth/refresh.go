package auth

import (
	"context"
	"time"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
)

const msgRefreshTokenInvalido = "refresh token inválido ou expirado"

// RefreshUseCase troca um refresh token válido por um novo par access+refresh,
// revogando o antigo (rotação real).
type RefreshUseCase struct {
	refreshTokens domainauth.RepositorioRefreshToken
	jwtAuth       *infraauth.AutenticadorJWT
	refreshTTL    time.Duration
}

func NewRefreshUseCase(refreshTokens domainauth.RepositorioRefreshToken, jwtAuth *infraauth.AutenticadorJWT, refreshTTL time.Duration) *RefreshUseCase {
	return &RefreshUseCase{refreshTokens: refreshTokens, jwtAuth: jwtAuth, refreshTTL: refreshTTL}
}

func (uc *RefreshUseCase) Executar(ctx context.Context, input RefreshInput) (RefreshOutput, error) {
	hash := infraauth.HashRefreshToken(input.RefreshTokenBruto)
	rt, err := uc.refreshTokens.BuscarPorHash(ctx, hash)
	if err != nil || !rt.EstaValido() {
		return RefreshOutput{}, shared.NewUnauthorizedError(msgRefreshTokenInvalido)
	}

	if err := uc.refreshTokens.Revogar(ctx, rt.ID); err != nil {
		return RefreshOutput{}, shared.NewInternalError("erro ao revogar refresh token", err)
	}

	novoPar, err := gerarTokens(ctx, uc.refreshTokens, uc.jwtAuth, rt.Tipo, rt.Papel, uc.refreshTTL, rt.UsuarioID)
	if err != nil {
		return RefreshOutput{}, err
	}

	return RefreshOutput{AccessToken: novoPar.AccessToken, RefreshToken: novoPar.RefreshToken}, nil
}
