package auth

import (
	"context"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// RefreshTokenRepository persiste e consulta refresh tokens.
// Implementado em internal/infrastructure/persistence/mysql/auth.
type RefreshTokenRepository interface {
	Salvar(ctx context.Context, rt *RefreshToken) error
	BuscarPorHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	Revogar(ctx context.Context, id uint64) error
	AccessTokenRevogado(ctx context.Context, jti string) (bool, error)
}

// CredenciaisRepository busca a credencial de login por email.
// Uma interface, duas implementações: usuario (interno) e cliente
// (injetadas no wiring por endpoint, não decididas em runtime).
type CredenciaisRepository interface {
	BuscarPorEmail(ctx context.Context, email string) (*Credencial, error)
}

// UsuarioStatusRepository consulta se um usuário (por ID) ainda está ativo.
// Usado por AuthenticationMiddleware pra rejeitar requisições de usuários
// inativados após o access token já ter sido emitido. Inativar só no login
// não bastaria, o token continuaria válido até expirar. Mesmo padrão de
// CredenciaisRepository: uma interface, implementações por tipo de usuário.
type UsuarioStatusRepository interface {
	EstaAtivo(ctx context.Context, id uint64) (bool, error)
}

// Credencial é a projeção mínima necessária pro fluxo de login.
type Credencial struct {
	ID                 uint64
	SenhaHash          string
	Papel              shared.PapelUsuario
	Ativo              bool
	RequerAlterarSenha bool
}
