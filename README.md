# Sistema de Oficina Mecânica: Tech Challenge Fase 1

API de gestão de ordens de serviço de uma oficina mecânica. Back-end em Go, Gin, GORM, MySQL 8.

Decisões e detalhes de arquitetura em [`docs/ARQUITETURA.md`](docs/ARQUITETURA.md) e nos [ADRs](docs/adr/).

## Pré-requisitos

- Go 1.26 ou superior
- Docker e Docker Compose
- `make`

## Configuração

Copie o `.env.example` para `.env` e ajuste o que precisar:

```bash
cp .env.example .env
```

| Variável | Descrição | Default |
|---|---|---|
| `APP_PORT` | Porta HTTP interna da API (usada pelo `config.go` e dentro do container) | `8080` |
| `APP_HOST_PORT` | Porta exposta no host para a API (só usada pelo `compose.yml`) | `8080` |
| `DB_HOST` | Host do MySQL (`localhost` local, sobrescrito para `mysql` dentro do compose) | `localhost` |
| `DB_PORT` | Porta interna do MySQL | `3306` |
| `DB_HOST_PORT` | Porta exposta no host para o MySQL (só usada pelo `compose.yml`) | `3306` |
| `DB_USER` | Usuário do MySQL | `root` |
| `DB_PASSWORD` | Senha do MySQL | `root` |
| `DB_NAME` | Nome do banco | `mecanica` |
| `DB_MAX_OPEN_CONNS` | Máximo de conexões abertas no pool | `25` |
| `DB_MAX_IDLE_CONNS` | Máximo de conexões ociosas mantidas no pool | `5` |
| `DB_CONN_MAX_LIFETIME_MINUTES` | Tempo máximo, em minutos, que uma conexão pode ficar aberta antes de ser reciclada | `5` |
| `SONAR_HOST_PORT` | Porta exposta no host para o SonarQube (só usada pelo `quality/compose.quality.yml`) | `9000` |
| `SONAR_TOKEN` | Token de autenticação do SonarQube, usado pelo serviço `sonar-scanner` (gerado na UI, ver seção SonarQube abaixo) | *(vazio)* |
| `SONAR_HOST_URL` | URL do SonarQube usada pelo `sonar-scanner` (nome do serviço na rede docker; só muda se você apontar pra um SonarQube externo) | `http://sonarqube:9000` |

## Comandos disponíveis

```bash
make help
```

Lista todos os goals do `Makefile` com uma descrição curta de cada um.

## Rodando via Docker

Fluxo recomendado para subir o projeto inteiro sem instalar nada além de Docker.

Sobe API e MySQL, com volume persistente para o banco:

```bash
make up
```

Confirma que a API está no ar:

```bash
curl http://localhost:8080/v1/health
# {"status":"ok"}
```

Derruba o ambiente, mantendo o volume do banco:

```bash
make down
```

## Rodando em modo dev local

Fluxo para quem vai editar código com hot reload e debugger, rodando a API fora do container e o MySQL dentro de um container isolado.

Instala as dependências do módulo e as ferramentas de dev (`swag`, `migrate`, `air`, `delve`):

```bash
make setup
```

As ferramentas vão para `$(go env GOPATH)/bin`; o `Makefile` já chama todas por caminho completo, não precisa desse diretório estar no `PATH` do seu shell.

`make run`, `make dev` e `make debug` carregam as variáveis do `.env` automaticamente (o `Makefile` faz `include .env` e exporta pro processo). Garanta que o `.env` existe (`cp .env.example .env`, seção Configuração acima) antes de rodar qualquer um desses.

Sobe só o MySQL:

```bash
make db-up
```

Aplica as migrations (schema versionado em `migrations/mysql/`, runner em `migrations/main.go`):

```bash
make migrate-up
```

`make db-setup` faz os dois passos de uma vez (`db-up` + `migrate-up`). Outros comandos: `make migrate-down` (desfaz a última), `make migrate-version` (mostra a versão atual) e `make migrate-force VERSION=N` (força a versão sem rodar SQL, só pra corrigir um estado `dirty`).

Cria um usuário interno direto no banco (`cmd/create-user`), sem passar pelo HTTP/JWT; resolve o bootstrap do primeiro admin (que precisaria de um admin já existente pra bater em `POST /v1/usuarios`) e serve pra testar o domínio manualmente:

```bash
make create-user NOME="Admin Oficina" EMAIL=admin@oficina.com SENHA=senha123
```

`PAPEL` é opcional, default `ADMINISTRADOR` (outros valores válidos: `MECANICO`, `ATENDENTE`). A senha nasce provisória (`requer_alterar_senha=true`), igual a qualquer usuário criado por um admin — troca obrigatória no primeiro login.

Roda a API com hot reload via `air`, que recompila e reinicia o processo a cada alteração em um arquivo `.go`:

```bash
make dev
```

Para parar o banco depois:

```bash
make db-down
```

### Debug com delve

Sobe a API sob o `dlv`, no modo headless, escutando na porta `2345`:

```bash
make debug
```

Pra debugar via VS Code, use a configuração já versionada em `.vscode/launch.json` ("Attach to Delve"). Com `make debug` rodando em um terminal, abra a aba Run and Debug no VS Code e inicie essa configuração: ela conecta na porta `2345` e permite pôr breakpoints, inspecionar variáveis e navegar no call stack normalmente.

Pra debugar via linha de comando, sem VS Code, conecte um client do delve na mesma porta em outro terminal:

```bash
dlv connect 127.0.0.1:2345
```

Isso abre o prompt interativo do delve (`break`, `continue`, `print`, etc.) contra o processo que já está rodando.

## Testes

```bash
make test
```

Roda os testes unitários de todos os pacotes com cobertura.

### Testes de integração

```bash
make test-integration
```

Sobem um MySQL real via `testcontainers` (precisa do Docker rodando), aplicam as migrations de produção e montam o `wiring.Container`/router reais; nada de mock, é o router completo batendo num banco de verdade. Ficam em `test/integration/`, atrás da build tag `integration` (por isso não entram em `make test`; container leva alguns segundos pra subir, não faz sentido no loop rápido do dia a dia).

Organização de `test/integration/`:
- `setup_test.go`: `TestMain`, sobe o container e monta a aplicação uma vez por execução do pacote.
- `fixtures_test.go`: dados de teste (`resetDB`, `seedUsuario`).
- `client_test.go`: helpers de request HTTP (`doRequest`, `doLogin`).
- Um arquivo por cenário de negócio (ciclo de auth completo, inativação em sessão ativa, senha provisória, autorização por papel, etc.).

### Cobertura

```bash
make coverage
```

Roda os testes com `-coverprofile`, imprime o total de cobertura no terminal e gera `test/reports/coverage.html`. Pra ver o relatório detalhado por linha, abra o arquivo gerado no navegador:

```bash
open test/reports/coverage.html
```

`test/reports/coverage.out` e `test/reports/coverage.html` são gitignored, ficam só localmente.

## Lint

```bash
make lint
```

## SonarQube (análise estática)

Sobe um SonarQube local (banco embarcado H2, só pra análise local mesmo) via `quality/compose.quality.yml`:

```bash
make sonar-up
```

Acesse `http://localhost:9000` (login inicial `admin`/`admin`, troca de senha obrigatória no primeiro acesso). Crie um projeto com key `soat-architecture` e gere um token em **My Account → Security → Generate Token**.

Em ambiente Linux, se o SonarQube não subir por causa de `vm.max_map_count`, ajuste no host:

```bash
sudo sysctl -w vm.max_map_count=524288
```

Cole o token gerado no `.env` (nunca no `.env.example`):

```bash
SONAR_TOKEN=<seu_token>
```

Rode a análise (gera cobertura via `make coverage` e escaneia via serviço `sonar-scanner` do `quality/compose.quality.yml`, na mesma rede docker do `sonarqube`):

```bash
make sonar-scan
```

Tanto `sonarqube` quanto `sonar-scanner` carregam variáveis do `.env` (`env_file`, mesmo padrão dos serviços `app`/`mysql`); não precisa passar nada na linha de comando além do `.env` configurado.

Configuração do projeto de análise fica em `sonar-project.properties` (project key, paths, exclusões, path do `test/reports/coverage.out`).

Pra parar o servidor:

```bash
make sonar-down
```

## Segurança (SCA/SAST/DAST)

Justificativa de cada ferramenta em [`docs/SECURITY.md`](docs/SECURITY.md).

```bash
make sec-sca    # govulncheck -> security/reports/govulncheck.json
make sec-sast   # gosec -> security/reports/gosec.json
make sec-dast   # OWASP ZAP -> security/reports/zap-report-{admin,atendente,mecanico,cliente}.{html,json}
```

`sec-dast` sobe uma stack Docker isolada (`security/compose.dast.yml`, projeto `soat-architecture-dast`: `app-dast` + `mysql-dast` com `tmpfs`) — não usa o `mysql`/`app` de desenvolvimento, então não suja o banco local com dados de teste. A stack some inteira (`down -v`) ao fim do script.

O scan roda autenticado uma vez por papel (admin, atendente, mecânico, cliente), provando que RBAC bloqueia de fato quem não tem permissão, não só que a rota responde. Falsos positivos conhecidos (achados em `/swagger`, documentados em `security/zap-rules.tsv`) são suprimidos automaticamente do relatório.

## Swagger

```bash
make swagger
```

Gera a documentação Swagger (via `swag`) a partir das anotações nos handlers, em `docs/swagger/` (`docs.go`, `swagger.json`, `swagger.yaml`). O `docs.go` é importado pelo módulo, então esses arquivos são versionados no repositório; rode `make swagger` de novo depois de anotar ou alterar um handler.

A UI fica em `/swagger/index.html`. Com a API no ar (`make up` ou `make dev`), acesse:

```
http://localhost:8080/swagger/index.html
```

## Erros de domínio

Erros de negócio (não encontrado, validação, conflito) trafegam da camada de domínio/aplicação até o HTTP como `*shared.AppError` (`internal/domain/shared/errors.go`):

```go
shared.NewNotFoundError("cliente não encontrado")
shared.NewValidationErrorWithDetails("dados inválidos", []string{"nome é obrigatório"})
shared.NewConflictError("ordem já finalizada")
shared.NewInternalError("erro ao consultar banco", err) // encapsula erro de infra
```

Regra pros use cases: erro de regra de negócio nasce como `*shared.AppError` na origem (domain ou application); erro de infra (repositório, driver) é encapsulado com `shared.NewInternalError("mensagem", err)` antes de subir. `fmt.Errorf("...: %w", err)` no meio do caminho não quebra nada, o mapper HTTP usa `errors.As` e enxerga através do wrap.

No handler, não monta `gin.H` na mão delega pro pacote de erro HTTP:

```go
if err != nil {
    httperror.RespondError(c, err)
    return
}
```

`RespondError` (`internal/infrastructure/http/httperror/errors.go`) traduz `Kind` pra status HTTP e devolve `{"type", "message", "details"}` (`httperror.ErrorResponse`). Há também um atalho por `Kind`: `RespondNotFoundError`, `RespondValidationError`, `RespondConflictError`, `RespondForbiddenError`, `RespondUnauthorizedError`, `RespondInternalError` pra handlers que já sabem o `Kind` sem montar um `*shared.AppError` primeiro:

| Kind           | Status | Atalho                       |
|----------------|--------|-------------------------------|
| `not_found`    | 404    | `RespondNotFoundError`         |
| `validation`   | 400    | `RespondValidationError`       |
| `conflict`     | 409    | `RespondConflictError`         |
| `forbidden`    | 403    | `RespondForbiddenError`        |
| `unauthorized` | 401    | `RespondUnauthorizedError`     |
| `internal`     | 500    | `RespondInternalError`         |
| `unavailable`  | 503    | `RespondUnavailableError`      |

Erro que não é `*shared.AppError` (ou `Kind` desconhecido) cai em 500 genérico mensagem interna nunca vaza pra resposta.

`httperror` é um pacote-folha deliberado: fica fora da raiz de `internal/infrastructure/http/` justamente pra domínios (`auth/`, `health/`) poderem importá-lo sem criar ciclo com `router.go`, que por sua vez importa os pacotes de domínio pra registrar rotas (ver ADR 0005).

`health/handler.go` usa `RespondUnavailableError` (`Kind` `unavailable`, 503) quando o ping no banco falha, mesmo sistema de erro dos demais domínios, sem exceção.

## Middlewares de autenticação/autorização

`internal/infrastructure/http/middleware/` tem dois middlewares Gin, pensados pra empilhar em sequência numa rota protegida.

### `AuthenticationMiddleware` valida o JWT

Lê o header `Authorization: Bearer <token>`, valida assinatura (HS256) e expiração via `AutenticadorJWT.ValidarAccessToken`, e injeta os claims (`*domainauth.AppClaims`) no `gin.Context` sob a chave `middleware.ClaimsContextKey`. Sem token, token malformado ou inválido → 401 (`httperror.RespondError` com `shared.NewUnauthorizedError`).

```go
middleware.AuthenticationMiddleware(c.JWTAuth) // c é *wiring.Container
```

Não decide quem pode acessar o quê, só garante "esse token é válido e é desse usuário". Isso é trabalho do próximo middleware.

### `AuthorizationMiddleware` valida o tipo de usuário

Lê o `*domainauth.AppClaims` que o `AuthenticationMiddleware` deixou no contexto e compara `claims.Tipo` com o `TipoUsuario` esperado pra rota. Tipo errado (ou claims ausente, ou `AuthenticationMiddleware` não rodou antes) → 403 (`httperror.RespondForbiddenError`).

```go
middleware.AuthorizationMiddleware(domainauth.TipoInterno) // ou domainauth.TipoCliente
```

**Sempre nessa ordem**, `AuthorizationMiddleware` depende do que `AuthenticationMiddleware` põe no contexto:

```go
rg.GET(
    "/ordens-servico",
    middleware.AuthenticationMiddleware(c.JWTAuth),
    middleware.AuthorizationMiddleware(domainauth.TipoInterno),
    handler.Listar,
)
```

### Adicionando numa rota nova

Em `Register<Dominio>Routes(rg *gin.RouterGroup, c *wiring.Container)` (ex. `internal/infrastructure/http/ordemservico/routes.go`), encadeie os middlewares antes do handler final, Gin aceita quantos `gin.HandlerFunc` forem passados, executados em ordem:

```go
func RegisterOrdemServicoRoutes(rg *gin.RouterGroup, c *wiring.Container) {
	os := rg.Group("/ordens-servico")
	os.Use(middleware.AuthenticationMiddleware(c.JWTAuth))

	os.GET("", c.OrdemServicoHandler.Listar) // qualquer tipo autenticado
	os.POST("",
		middleware.AuthorizationMiddleware(domainauth.TipoInterno), // só interno
		c.OrdemServicoHandler.Criar,
	)
}
```

`rg.Use(...)` aplica o middleware a todas as rotas daquele grupo; middleware passado direto no verbo HTTP (`os.POST("", middleware..., handler)`) vale só pra aquela rota. Rota sem nenhum dos dois fica pública (ex. `/v1/health`, `/v1/auth/login`).

## Mocks

```bash
make mocks
```

Gera mocks (via `mockery`, configurado em `.mockery.yaml`) para as interfaces de `internal/domain/` e `internal/application/`, com o padrão `with-expecter` do `testify/mock`. Cada mock nasce em `mocks/` dentro do pacote da interface, por exemplo `internal/domain/ordemservico/mocks/Repository.go`. Os arquivos gerados são versionados no repositório; rode `make mocks` de novo sempre que uma interface mudar.

## Git hooks

O repositório tem um stub de hook `pre-push` em `.dev/hooks/stubs/pre-push.stub`, que roda `make mocks`, `make lint`, `make test` e `make swagger` antes de liberar o push. Ele não vem instalado por padrão (hooks do Git não são versionados dentro de `.git/`); cada dev instala localmente:

```bash
make hooks-install
```

Copia o stub para `.git/hooks/pre-push` e dá permissão de execução. Se o push falhar por causa do hook, corrija o que ele acusou (mocks desatualizados, lint, teste quebrado, swagger desatualizado) e tente de novo.

Para remover:

```bash
make hooks-uninstall
```

## Empacotamento com vendor

```bash
make vendor
```

Roda `go mod tidy` e depois `go mod vendor`. `vendor/` é gitignored, serve só como cache local opcional.
