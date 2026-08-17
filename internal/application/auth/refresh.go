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
	refreshTokens domainauth.RefreshTokenRepository
	jwtAuth       domainauth.JWTProvider
	accessTTL     time.Duration
	refreshTTL    time.Duration
}

// NewRefreshUseCase monta o use case com o repositório de refresh tokens e o
// autenticador JWT, compartilhados entre interno e cliente (o token antigo já
// carrega o TipoUsuario, propagado pro novo par).
func NewRefreshUseCase(refreshTokens domainauth.RefreshTokenRepository, jwtAuth domainauth.JWTProvider, accessTTL time.Duration, refreshTTL time.Duration) *RefreshUseCase {
	return &RefreshUseCase{refreshTokens: refreshTokens, jwtAuth: jwtAuth, accessTTL: accessTTL, refreshTTL: refreshTTL}
}

// Executar valida o refresh token bruto (existe, não expirado, não revogado),
// revoga-o e emite um novo par access+refresh, rotação real, o token antigo
// nunca é reaproveitado mesmo que a chamada seguinte falhe antes de persistir
// o novo par.
func (uc *RefreshUseCase) Executar(ctx context.Context, input RefreshInput) (RefreshOutput, error) {
	hash := infraauth.HashRefreshToken(input.RefreshTokenBruto)
	rt, err := uc.refreshTokens.BuscarPorHash(ctx, hash)
	if err != nil || !rt.EstaValido() {
		return RefreshOutput{}, shared.NewUnauthorizedError(msgRefreshTokenInvalido)
	}

	if err := uc.refreshTokens.Revogar(ctx, rt.ID); err != nil {
		return RefreshOutput{}, shared.NewInternalError("erro ao revogar refresh token", err)
	}

	novoPar, err := gerarTokens(ctx, uc.refreshTokens, uc.jwtAuth, rt.Tipo, rt.Papel, uc.accessTTL, uc.refreshTTL, rt.UsuarioID)
	if err != nil {
		return RefreshOutput{}, err
	}

	return RefreshOutput{
		AccessToken:           novoPar.AccessToken,
		RefreshToken:          novoPar.RefreshToken,
		AccessTokenExpiresIn:  novoPar.AccessTokenExpiresIn,
		RefreshTokenExpiresIn: novoPar.RefreshTokenExpiresIn,
	}, nil
}
