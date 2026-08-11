# Wiring e registro de rotas organizados

## Contexto

Hoje `cmd/api/main.go` monta config e conexão de banco, e `internal/infrastructure/http/router.go`
constrói o `*gin.Engine` e registra rotas diretamente, tudo em um único lugar. Só existe o domínio
`ordemservico` com esqueleto (use case e DTO ainda "Implementação pendente") e uma rota real: `/v1/health`.

À medida que novos domínios (cliente, veiculo, peca, servico) ganharem use cases e handlers, `router.go`
cresceria misturando composição de dependências com registro de rotas. O objetivo é introduzir uma
estrutura de wiring (injeção de dependências) e um padrão de registro de rotas por domínio, sem
implementar handlers que ainda não existem.

## Escopo

- Cobre apenas o que já existe: health check.
- Não cria stubs de wiring/rotas para cliente, veiculo, peca, servico (aguardam use cases reais).
- Estabelece o padrão a ser repetido quando esses domínios forem implementados.

## Design

### Container (wiring)

Novo pacote `internal/infrastructure/wiring`, arquivo `container.go`:

```go
package wiring

type Container struct {
    Config *config.Config
    DB     *gorm.DB
}

func NewContainer(cfg *config.Config, db *gorm.DB) *Container {
    return &Container{Config: cfg, DB: db}
}
```

Responsabilidade: compor as dependências compartilhadas da aplicação (config, conexão de banco e,
futuramente, repositórios e use cases por domínio). É o único lugar que monta o grafo de dependências.

### Registro de rotas por domínio

Um arquivo de rotas por domínio dentro do pacote `http` (sem subpacote `routes` — YAGNI enquanto só
existe um domínio). Para health: `internal/infrastructure/http/health_routes.go`:

```go
package http

func RegisterHealthRoutes(rg *gin.RouterGroup, c *wiring.Container) {
    rg.GET("/health", NewHealthHandler(c.DB))
}
```

Padrão a repetir: `<dominio>_routes.go` com `Register<Dominio>Routes(rg *gin.RouterGroup, c *wiring.Container)`.

### Router

`router.go` deixa de montar dependências e passa a apenas configurar o engine Gin e delegar registro
de rotas para as funções `Register*Routes`:

```go
func NewRouter(c *wiring.Container) *gin.Engine {
    router := gin.New()
    router.Use(gin.Logger())
    router.Use(gin.Recovery())

    router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

    v1 := router.Group("/v1")
    RegisterHealthRoutes(v1, c)

    return router
}
```

### main.go

```go
container := wiring.NewContainer(cfg, db)
router := httpinfra.NewRouter(container)
```

## Extensão futura

Quando um domínio (ex: `cliente`) ganhar repositório e use case reais:

1. Adicionar campos no `Container` (ex: `ClienteRepository`, `CadastrarClienteUseCase`) e construí-los
   em `NewContainer`.
2. Criar `cliente_routes.go` com `RegisterClienteRoutes(rg *gin.RouterGroup, c *wiring.Container)`.
3. Chamar `RegisterClienteRoutes(v1, c)` em `router.go`.

Nenhuma mudança estrutural adicional é necessária — o padrão já comporta N domínios.

## Testes

- `go build ./...` e `go vet ./...` devem passar.
- Testes de integração existentes em `test/integration` (se cobrirem `/v1/health`) devem continuar passando.
