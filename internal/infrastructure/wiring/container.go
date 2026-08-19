package wiring

import (
	"time"

	"gorm.io/gorm"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	mysqlauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/auth"
	mysqlpeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/peca"
	mysqlusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/usuario"
)

// Container compõe as dependências compartilhadas da aplicação
// (config, conexão de banco e repositórios e use cases por domínio).
type Container struct {
	Config *config.Config
	DB     *gorm.DB

	JWTAuth           *infraauth.AuthenticatorJWT
	RefreshTokensRepo domainauth.RefreshTokenRepository
	UsuarioRepo       domainusuario.UsuarioRepository
	UsuarioStatusRepo domainauth.UsuarioStatusRepository
	PecaRepo          domainpeca.Repository

	LoginInternoUC *appauth.LoginUseCase
	RefreshUC      *appauth.RefreshUseCase
	LogoutUC       *appauth.LogoutUseCase

	CriarUsuarioUC        *appusuario.CriarUsuarioUseCase
	AtualizarUsuarioUC    *appusuario.AtualizarUsuarioUseCase
	AlterarSenhaUC        *appusuario.AlterarSenhaUseCase
	AtivarUsuarioUC       *appusuario.AtivarUsuarioUseCase
	InativarUsuarioUC     *appusuario.InativarUsuarioUseCase
	BuscarUsuarioLogadoUC *appusuario.BuscarUsuarioLogadoUseCase

	CadastrarPecaUC      *apppeca.CadastrarPecaUseCase
	AtualizarPecaUC      *apppeca.AtualizarPecaUseCase
	AtivarPecaUC         *apppeca.AtivarPecaUseCase
	InativarPecaUC       *apppeca.InativarPecaUseCase
	ConsultarPecaPorIDUC *apppeca.ConsultarPecaPorIDUseCase
	ReporEstoquePecaUC   *apppeca.ReporEstoqueUseCase
}

// NewContainer monta o grafo de dependências da aplicação.
func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	c := &Container{Config: cfg, DB: db}
	c.RefreshTokensRepo = mysqlauth.NewRefreshTokenRepository(db)
	c.UsuarioRepo = mysqlusuario.NewUsuarioRepository(db)
	c.PecaRepo = mysqlpeca.NewRepository(db)
	if cfg == nil {
		return c
	}

	accessTTL := time.Duration(cfg.JWTAccessTokenTTLMinutes) * time.Minute
	refreshTTL := time.Duration(cfg.JWTRefreshTokenTTLHours) * time.Hour
	c.JWTAuth = infraauth.NewAuthenticatorJWT(cfg.JWTSecret, accessTTL)

	credenciaisInterno := mysqlusuario.NewCredenciaisAdapter(c.UsuarioRepo)
	c.UsuarioStatusRepo = credenciaisInterno
	c.LoginInternoUC = appauth.NewLoginUseCase(credenciaisInterno, c.RefreshTokensRepo, c.JWTAuth, domainauth.TipoInterno, accessTTL, refreshTTL)
	c.RefreshUC = appauth.NewRefreshUseCase(c.RefreshTokensRepo, c.JWTAuth, accessTTL, refreshTTL)
	c.LogoutUC = appauth.NewLogoutUseCase(c.RefreshTokensRepo)

	c.CriarUsuarioUC = appusuario.NewCriarUsuarioUseCase(c.UsuarioRepo)
	c.AtualizarUsuarioUC = appusuario.NewAtualizarUsuarioUseCase(c.UsuarioRepo)
	c.AlterarSenhaUC = appusuario.NewAlterarSenhaUseCase(c.UsuarioRepo)
	c.AtivarUsuarioUC = appusuario.NewAtivarUsuarioUseCase(c.UsuarioRepo)
	c.InativarUsuarioUC = appusuario.NewInativarUsuarioUseCase(c.UsuarioRepo)
	c.BuscarUsuarioLogadoUC = appusuario.NewBuscarUsuarioLogadoUseCase(c.UsuarioRepo)

	c.CadastrarPecaUC = apppeca.NewCadastrarPecaUseCase(c.PecaRepo)
	c.AtualizarPecaUC = apppeca.NewAtualizarPecaUseCase(c.PecaRepo)
	c.AtivarPecaUC = apppeca.NewAtivarPecaUseCase(c.PecaRepo)
	c.InativarPecaUC = apppeca.NewInativarPecaUseCase(c.PecaRepo)
	c.ConsultarPecaPorIDUC = apppeca.NewConsultarPecaPorIDUseCase(c.PecaRepo)
	c.ReporEstoquePecaUC = apppeca.NewReporEstoqueUseCase(c.PecaRepo)

	return c
}
