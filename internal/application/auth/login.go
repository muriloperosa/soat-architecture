package auth

import (
	"context"
	"time"

	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"golang.org/x/crypto/bcrypt"
)

const msgCredenciaisInvalidas = "credenciais inválidas"

// LoginUseCase autentica por email+senha e emite o par access+refresh token.
// A fonte de credenciais (interno ou cliente) e o TipoUsuario já vêm decididos
// no wiring, não em runtime.
type LoginUseCase struct {
	credenciais   domainauth.RepositorioCredenciais
	refreshTokens domainauth.RepositorioRefreshToken
	jwtAuth       domainauth.JWTProvider
	tipo          domainauth.TipoUsuario
	refreshTTL    time.Duration
}

// NewLoginUseCase monta o use case com a fonte de credenciais e o TipoUsuario
// já decididos (injetados pelo wiring, um por endpoint de login).
func NewLoginUseCase(
	credenciais domainauth.RepositorioCredenciais,
	refreshTokens domainauth.RepositorioRefreshToken,
	jwtAuth domainauth.JWTProvider,
	tipo domainauth.TipoUsuario,
	refreshTTL time.Duration,
) *LoginUseCase {
	return &LoginUseCase{
		credenciais:   credenciais,
		refreshTokens: refreshTokens,
		jwtAuth:       jwtAuth,
		tipo:          tipo,
		refreshTTL:    refreshTTL,
	}
}

// Executar valida email+senha contra a credencial cadastrada e, se válida,
// emite um novo par access+refresh token. Credencial inexistente e senha
// errada retornam o mesmo erro genérico.
func (uc *LoginUseCase) Executar(ctx context.Context, input LoginInput) (LoginOutput, error) {
	cred, err := uc.credenciais.BuscarPorEmail(ctx, input.Email)
	if err != nil {
		return LoginOutput{}, shared.NewUnauthorizedError(msgCredenciaisInvalidas)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(cred.SenhaHash), []byte(input.Senha)); err != nil {
		return LoginOutput{}, shared.NewUnauthorizedError(msgCredenciaisInvalidas)
	}

	return gerarTokens(ctx, uc.refreshTokens, uc.jwtAuth, uc.tipo, cred.Papel, uc.refreshTTL, cred.ID)
}

// gerarTokens gera e persiste um novo par access+refresh token pro usuário.
// Reusado pelo RefreshUseCase na rotação.
func gerarTokens(ctx context.Context, refreshTokens domainauth.RepositorioRefreshToken, jwtAuth domainauth.JWTProvider, tipo domainauth.TipoUsuario, papel domainauth.PapelUsuario, refreshTTL time.Duration, usuarioID string) (LoginOutput, error) {
	accessToken, err := jwtAuth.GerarAccessToken(usuarioID, tipo, papel)
	if err != nil {
		return LoginOutput{}, shared.NewInternalError("erro ao gerar access token", err)
	}

	refreshBruto, err := jwtAuth.GerarRefreshTokenBruto()
	if err != nil {
		return LoginOutput{}, shared.NewInternalError("erro ao gerar refresh token", err)
	}

	rt := &domainauth.RefreshToken{
		UsuarioID: usuarioID,
		Tipo:      tipo,
		Papel:     papel,
		TokenHash: infraauth.HashRefreshToken(refreshBruto),
		ExpiraEm:  time.Now().Add(refreshTTL),
	}
	if err := refreshTokens.Salvar(ctx, rt); err != nil {
		return LoginOutput{}, shared.NewInternalError("erro ao salvar refresh token", err)
	}

	return LoginOutput{AccessToken: accessToken, RefreshToken: refreshBruto}, nil
}
