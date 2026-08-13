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

Roda os testes unitários de todos os pacotes com cobertura. Testes de integração, MySQL real via `testcontainers`, ficam em `test/integration/`.

### Cobertura

```bash
make coverage
```

Roda os testes com `-coverprofile`, imprime o total de cobertura no terminal e gera `coverage.html`. Pra ver o relatório detalhado por linha, abra o arquivo gerado no navegador:

```bash
open coverage.html
```

`coverage.out` e `coverage.html` são gitignored, ficam só localmente.

## Lint

```bash
make lint
```

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
shared.NewValidationErrorComDetails("dados inválidos", []string{"nome é obrigatório"})
shared.NewConflictError("ordem já finalizada")
shared.NewInternalError("erro ao consultar banco", err) // encapsula erro de infra
```

Regra pros use cases: erro de regra de negócio nasce como `*shared.AppError` na origem (domain ou application); erro de infra (repositório, driver) é encapsulado com `shared.NewInternalError("mensagem", err)` antes de subir. `fmt.Errorf("...: %w", err)` no meio do caminho não quebra nada — o mapper HTTP usa `errors.As` e enxerga através do wrap.

No handler, não monta `gin.H` na mão — delega pro mapper:

```go
if err != nil {
    http.RespondError(c, err)
    return
}
```

`RespondError` (`internal/infrastructure/http/errors.go`) traduz `Kind` pra status HTTP e devolve `{"type", "message", "details"}`:

| Kind         | Status |
|--------------|--------|
| `not_found`  | 404    |
| `validation` | 400    |
| `conflict`   | 409    |
| `internal`   | 500    |

Erro que não é `*shared.AppError` (ou `Kind` desconhecido) cai em 500 genérico — mensagem interna nunca vaza pra resposta.

`health_handler.go` é exceção: é probe de infra (ping no banco), não erro de domínio, então continua montando a resposta 503 direto.

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
