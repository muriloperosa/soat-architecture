package wiring

import (
	"time"

	"gorm.io/gorm"

	appauth "github.com/muriloperosa/soat-architecture/internal/application/auth"
	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
	domainveiculo "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	infraauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	mysqlauth "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/auth"
	mysqlcliente "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/cliente"
	mysqlpeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/peca"
	mysqlusuario "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/usuario"
	mysqlveiculo "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/veiculo"
)

type Container struct {
	Config *config.Config
	DB     *gorm.DB

	JWTAuth           *infraauth.AuthenticatorJWT
	RefreshTokensRepo domainauth.RefreshTokenRepository
	UsuarioRepo       domainusuario.UsuarioRepository
	UsuarioStatusRepo domainauth.UsuarioStatusRepository
	ClienteRepository domaincliente.ClienteRepository
	ClienteStatusRepo domainauth.UsuarioStatusRepository
	PecaRepo          domainpeca.Repository
	VeiculoRepo       domainveiculo.Repository

	LoginInternoUC *appauth.LoginUseCase
	LoginClienteUC *appauth.LoginUseCase
	RefreshUC      *appauth.RefreshUseCase
	LogoutUC       *appauth.LogoutUseCase

	CriarUsuarioUC        *appusuario.CriarUsuarioUseCase
	AtualizarUsuarioUC    *appusuario.AtualizarUsuarioUseCase
	AlterarSenhaUC        *appusuario.AlterarSenhaUseCase
	AtivarUsuarioUC       *appusuario.AtivarUsuarioUseCase
	InativarUsuarioUC     *appusuario.InativarUsuarioUseCase
	BuscarUsuarioLogadoUC *appusuario.BuscarUsuarioLogadoUseCase

	CriarClienteUseCase                 *appcliente.CriarClienteUseCase
	AtualizarClienteUseCase             *appcliente.AtualizarClienteUseCase
	ConsultarClientePorIDUseCase        *appcliente.ConsultarClientePorIDUseCase
	ConsultarClientePorDocumentoUseCase *appcliente.ConsultarClientePorDocumentoUseCase
	AtivarClienteUseCase                *appcliente.AtivarClienteUseCase
	InativarClienteUseCase              *appcliente.InativarClienteUseCase
	AlterarSenhaClienteUseCase          *appcliente.AlterarSenhaClienteUseCase

	CadastrarPecaUC      *apppeca.CadastrarPecaUseCase
	AtualizarPecaUC      *apppeca.AtualizarPecaUseCase
	AtivarPecaUC         *apppeca.AtivarPecaUseCase
	InativarPecaUC       *apppeca.InativarPecaUseCase
	ConsultarPecaPorIDUC *apppeca.ConsultarPecaPorIDUseCase
	ReporEstoquePecaUC   *apppeca.ReporEstoqueUseCase

	CadastrarVeiculoUC         *appveiculo.CadastrarVeiculoUseCase
	AtualizarVeiculoUC         *appveiculo.AtualizarVeiculoUseCase
	AtivarVeiculoUC            *appveiculo.AtivarVeiculoUseCase
	InativarVeiculoUC          *appveiculo.InativarVeiculoUseCase
	ConsultarVeiculoPorIDUC    *appveiculo.ConsultarVeiculoPorIDUseCase
	ConsultarVeiculoPorPlacaUC *appveiculo.ConsultarVeiculoPorPlacaUseCase
}

func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
	c := &Container{Config: cfg, DB: db}
	c.RefreshTokensRepo = mysqlauth.NewRefreshTokenRepository(db)
	c.UsuarioRepo = mysqlusuario.NewUsuarioRepository(db)
	c.ClienteRepository = mysqlcliente.NewClienteRepository(db)
	c.PecaRepo = mysqlpeca.NewRepository(db)
	c.VeiculoRepo = mysqlveiculo.NewRepository(db)
	if cfg == nil {
		return c
	}

	accessTTL := time.Duration(cfg.JWTAccessTokenTTLMinutes) * time.Minute
	refreshTTL := time.Duration(cfg.JWTRefreshTokenTTLHours) * time.Hour
	c.JWTAuth = infraauth.NewAuthenticatorJWT(cfg.JWTSecret, accessTTL)

	credenciaisInterno := mysqlusuario.NewCredenciaisAdapter(c.UsuarioRepo)
	credenciaisCliente := mysqlcliente.NewCredenciaisAdapter(c.ClienteRepository)
	c.UsuarioStatusRepo = credenciaisInterno
	c.ClienteStatusRepo = credenciaisCliente
	c.LoginInternoUC = appauth.NewLoginUseCase(credenciaisInterno, c.RefreshTokensRepo, c.JWTAuth, domainauth.TipoInterno, accessTTL, refreshTTL)
	c.LoginClienteUC = appauth.NewLoginUseCase(credenciaisCliente, c.RefreshTokensRepo, c.JWTAuth, domainauth.TipoCliente, accessTTL, refreshTTL)
	c.RefreshUC = appauth.NewRefreshUseCase(c.RefreshTokensRepo, c.JWTAuth, accessTTL, refreshTTL)
	c.LogoutUC = appauth.NewLogoutUseCase(c.RefreshTokensRepo)

	c.CriarUsuarioUC = appusuario.NewCriarUsuarioUseCase(c.UsuarioRepo)
	c.AtualizarUsuarioUC = appusuario.NewAtualizarUsuarioUseCase(c.UsuarioRepo)
	c.AlterarSenhaUC = appusuario.NewAlterarSenhaUseCase(c.UsuarioRepo)
	c.AtivarUsuarioUC = appusuario.NewAtivarUsuarioUseCase(c.UsuarioRepo)
	c.InativarUsuarioUC = appusuario.NewInativarUsuarioUseCase(c.UsuarioRepo)
	c.BuscarUsuarioLogadoUC = appusuario.NewBuscarUsuarioLogadoUseCase(c.UsuarioRepo)

	c.CriarClienteUseCase = appcliente.NewCriarClienteUseCase(c.ClienteRepository)
	c.AtualizarClienteUseCase = appcliente.NewAtualizarClienteUseCase(c.ClienteRepository)
	c.ConsultarClientePorIDUseCase = appcliente.NewConsultarClientePorIDUseCase(c.ClienteRepository)
	c.ConsultarClientePorDocumentoUseCase = appcliente.NewConsultarClientePorDocumentoUseCase(c.ClienteRepository)
	c.AtivarClienteUseCase = appcliente.NewAtivarClienteUseCase(c.ClienteRepository)
	c.InativarClienteUseCase = appcliente.NewInativarClienteUseCase(c.ClienteRepository)
	c.AlterarSenhaClienteUseCase = appcliente.NewAlterarSenhaClienteUseCase(c.ClienteRepository)

	c.CadastrarPecaUC = apppeca.NewCadastrarPecaUseCase(c.PecaRepo)
	c.AtualizarPecaUC = apppeca.NewAtualizarPecaUseCase(c.PecaRepo)
	c.AtivarPecaUC = apppeca.NewAtivarPecaUseCase(c.PecaRepo)
	c.InativarPecaUC = apppeca.NewInativarPecaUseCase(c.PecaRepo)
	c.ConsultarPecaPorIDUC = apppeca.NewConsultarPecaPorIDUseCase(c.PecaRepo)
	c.ReporEstoquePecaUC = apppeca.NewReporEstoqueUseCase(c.PecaRepo)

	c.CadastrarVeiculoUC = appveiculo.NewCadastrarVeiculoUseCase(c.VeiculoRepo)
	c.AtualizarVeiculoUC = appveiculo.NewAtualizarVeiculoUseCase(c.VeiculoRepo)
	c.AtivarVeiculoUC = appveiculo.NewAtivarVeiculoUseCase(c.VeiculoRepo)
	c.InativarVeiculoUC = appveiculo.NewInativarVeiculoUseCase(c.VeiculoRepo)
	c.ConsultarVeiculoPorIDUC = appveiculo.NewConsultarVeiculoPorIDUseCase(c.VeiculoRepo)
	c.ConsultarVeiculoPorPlacaUC = appveiculo.NewConsultarVeiculoPorPlacaUseCase(c.VeiculoRepo)

	return c
}
