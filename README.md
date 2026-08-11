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
curl http://localhost:8080/health
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

As ferramentas vão para `$(go env GOPATH)/bin`. Garanta que esse caminho está no `PATH` do seu shell.

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

## Mocks

```bash
make mocks
```

Gera mocks (via `mockery`, configurado em `.mockery.yaml`) para as interfaces de `internal/domain/` e `internal/application/`, com o padrão `with-expecter` do `testify/mock`. Cada mock nasce em `mocks/` dentro do pacote da interface, por exemplo `internal/domain/ordemservico/mocks/Repository.go`. Os arquivos gerados são versionados no repositório; rode `make mocks` de novo sempre que uma interface mudar.

## Empacotamento com vendor

```bash
make vendor
```

Roda `go mod tidy` e depois `go mod vendor`. `vendor/` é gitignored, serve só como cache local opcional.