package wiring

import (
	"time"

	"gorm.io/gorm"

	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
)

// Container compõe as dependências compartilhadas da aplicação
// (config, conexão de banco e, futuramente, repositórios e use cases por domínio).
type Container struct {
	Config *config.Config
	DB     *gorm.DB

	JWTAuth *infraauth.AutenticadorJWT

	// TODO(Task 14): AuthInternoHandler *httphandler.AuthInternoHandler e
	// AuthClienteHandler *httphandler.AuthClienteHandler ficam bloqueados até
	// usuariointerno.Usuario e cliente.SenhaHash existirem — só então dá pra
	// montar CredenciaisRepository dos dois tipos (mysqlusuariointerno,
	// mysqlcliente) e os use cases de login/refresh/logout em cima deles.
	// RefreshTokenRepository (Task 13) já está pronto em
	// internal/infrastructure/persistence/mysql/auth, só falta plugar aqui
	// junto com os handlers.
}

// NewContainer monta o grafo de dependências da aplicação.
func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	c := &Container{Config: cfg, DB: db}
	if cfg == nil {
		return c
	}

	accessTTL := time.Duration(cfg.JWTAccessTokenTTLMinutes) * time.Minute
	c.JWTAuth = infraauth.NewAuthenticatorJWT(cfg.JWTSecret, accessTTL)

	return c
}
