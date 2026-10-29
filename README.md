# Sistema de Oficina Mecânica: Tech Challenge Fase 1

API de gestão de ordens de serviço de uma oficina mecânica, cobrindo o fluxo completo desde o cadastro do cliente e do veículo até a execução do serviço, passando por diagnóstico, orçamento, aprovação, reserva de peças e entrega. Back-end em Go, com Gin, GORM e MySQL 8.

## Overview de Arquitetura

O projeto segue Clean Architecture com núcleo tático de DDD, num único deployável. Domínio (entidades e value objects), aplicação (casos de uso) e infraestrutura (adaptadores HTTP e de persistência, frameworks) são pacotes separados, e as dependências apontam só para dentro: infraestrutura conhece aplicação e domínio, aplicação conhece só domínio, e domínio não conhece nada do projeto. Os repositórios e serviços externos que o domínio usa são interfaces declaradas nele e implementadas na infraestrutura. Essa regra é verificada por um teste automático (`test/architecture/`). Cada agregado (ordem de serviço, cliente, veículo, peça, usuário) carrega suas próprias regras de negócio em Value Objects auto validáveis, como `Documento`, `Placa` e `Status`, que nascem válidos ou não nascem.

A persistência usa GORM, mas de forma isolada: os models do GORM nunca vazam para a camada de domínio, existe um mapeamento explícito entre entidade e model na fronteira do repositório. Autenticação é feita via JWT, com um middleware que reavalida se o usuário segue ativo a cada requisição, não só no login, e um segundo middleware que autoriza por tipo e papel de usuário. Situações que exigem consistência entre agregados diferentes, como reservar peça e verificar estoque, passam por um `TransactionRunner` que garante atomicidade e usa `SELECT ... FOR UPDATE` para serializar concorrência.

O versionamento do schema do banco é feito com `golang-migrate`, migrations em `migrations/mysql/`, aplicadas via `make migrate-up`. As decisões estruturais do projeto (Clean Architecture, MySQL, GORM sem vazamento, framework Gin, convenção de nomes) estão registradas como ADRs em [`docs/adr/`](docs/adr/), e o detalhamento das convenções internas está em [`docs/ARQUITETURA.md`](docs/ARQUITETURA.md).

## Documentação

- [Guia de desenvolvimento](docs/DEVELOPMENT_GUIDE.md): pacote completo das informações necessárias para o desenvolvimento do projeto, serve como guia para os desenvolvedores
- [Arquitetura](docs/ARQUITETURA.md): convenções internas de camadas, mapeamento e injeção de dependência
- [ADRs](docs/adr/): decisões estruturais registradas e justificadas
- [Modelo de domínio](docs/DOMAIN_MODEL.md): agregados, entidades e Value Objects
- [Segurança](docs/SECURITY.md): ferramentas de SCA, SAST e DAST usadas e a justificativa de cada uma
- [Kubernetes](docs/KUBERNETES.md): manifestos em `k8s/`, requisitos da máquina, deploy no Minikube, comandos `kubectl` e demonstração do autoscaling
- [Swagger](docs/swagger/): documentação da API gerada a partir dos handlers (UI disponível em `/swagger/index.html` com a aplicação no ar)
- [Coleção Postman](.dev/postman/): coleção pronta para testar os endpoints da API

## Qualidade

O projeto mantém uma meta de cobertura de 80% ou mais nos domínios críticos, com testes unitários isolados por camada via mocks (`mockery`) e testes de integração que sobem um MySQL real com `testcontainers`, aplicam as migrations de produção e batem no router completo da aplicação, sem nenhum tipo de mock. `make lint` roda `go vet` e as regras estáticas do projeto antes de qualquer commit ser aceito.

A análise de vulnerabilidades cobre três frentes independentes: `govulncheck` para dependências conhecidas (SCA), `gosec` para padrões de risco no código fonte (SAST) e OWASP ZAP para ataques simulados contra a API em execução (DAST), este último rodando autenticado uma vez por papel de usuário para provar que o controle de acesso bloqueia de fato quem não tem permissão. SonarQube fica disponível localmente para análise estática complementar. Todos os detalhes e a justificativa de cada ferramenta estão em [`docs/SECURITY.md`](docs/SECURITY.md).

Um hook de `pre-push` local, instalado via `make hooks-install`, roda mocks, lint, testes e geração do Swagger antes de liberar qualquer push, evitando que código desatualizado ou quebrado chegue ao repositório remoto.

## Executar Localmente

Pré-requisitos: Docker e Docker Compose.

Sobe o ambiente inteiro com um único comando, sem exigir Go instalado localmente:

```bash
chmod +x .dev/scripts/setup.sh
.dev/scripts/setup.sh
```

- limpa containers e volumes de execuções anteriores
- builda as imagens da aplicação e do toolbox
- sobe o MySQL
- roda migrations, testes unitários e de integração
- popula o banco com usuários de login e dados fictícios
- sobe a API
- ao final, imprime as credenciais dos usuários seedados (admin, atendente, mecânico, cliente) para testar a API

## Deploy em Kubernetes

Os manifestos ficam em [`k8s/`](k8s/), organizados com Kustomize: uma base comum (Deployment, Service, ConfigMap e HPA da API) e overlays para Minikube e EKS. As migrations rodam num Job separado ([ADR 0006](docs/adr/0006-migrations-em-job-kubernetes.md)).

Pré-requisitos: Docker rodando, Minikube, `kubectl`, k6, Go 1.26 e `make`. Versões mínimas e recursos de máquina estão em [`docs/KUBERNETES.md`](docs/KUBERNETES.md#requisitos-da-máquina).

### Roteiro: subir no Minikube e demonstrar o autoscaling

**1. Criar o cluster.** Só na primeira vez ou depois de um `minikube delete`. Se ele só estava parado, basta `minikube start`.

```bash
minikube start --cpus=4 --memory=6g --disk-size=20g
minikube addons enable metrics-server
```

**2. Criar o arquivo de segredos.** Ele não vai para o git. Troque os dois `changeme`; a `DB_PASSWORD` deve ter só letras e números.

```bash
cp k8s/overlays/minikube/secrets.env.example k8s/overlays/minikube/secrets.env
```

**3. Construir a imagem e fazer o deploy.** O `k8s-deploy` aplica a configuração, roda as migrations num Job e sobe a API.

```bash
make k8s-build
make k8s-deploy
make k8s-status    # 2 pods da API em Running e o HPA com cpu: X%/70%
```

**4. Popular o banco.** Só quando o banco está vazio. Num terminal, abra o túnel até o MySQL e deixe rodando:

```bash
kubectl --context minikube -n oficina port-forward pod/oficina-mysql-0 3307:3306
```

Em outro terminal, rode o seed com os valores do `secrets.env`. Ele cria o usuário `admin@oficina.com` com a senha `Senha@123`:

```bash
DB_HOST=127.0.0.1 DB_PORT=3307 DB_USER=root DB_NAME=mecanica \
DB_PASSWORD=<valor do secrets.env> JWT_SECRET=<valor do secrets.env> \
go run ./cmd/seed
```

Depois feche o túnel.

**5. Obter a URL da API.** No macOS (driver Docker) o comando abre um túnel e precisa ficar rodando num terminal próprio. Anote a URL impressa.

```bash
make k8s-url
```

**6. Abrir o monitoramento**, um comando por terminal:

```bash
kubectl --context minikube -n oficina get hpa oficina-api -w
kubectl --context minikube -n oficina get pods -l app.kubernetes.io/name=oficina-api -w
watch -n 5 kubectl --context minikube -n oficina top pods    # no macOS: brew install watch
```

**7. Rodar o teste de carga.** São 4 minutos, em rampa até 100 usuários virtuais.

```bash
make load-test BASE_URL=<url do passo 5> K6_EMAIL=admin@oficina.com K6_SENHA=Senha@123
```

**8. Acompanhar a escala.** Cerca de um minuto depois do início a CPU passa da meta de 70% e o HPA sobe as réplicas de uma em uma, até 4 ou 5. Depois que o k6 termina, o número volta para 2 em mais ou menos um minuto e meio. As métricas do Minikube atualizam a cada 60 segundos, então a coluna `TARGETS` muda aos saltos. O motivo de cada decisão aparece em:

```bash
kubectl --context minikube -n oficina describe hpa oficina-api | grep SuccessfulRescale
```

**9. Encerrar.** `minikube stop` desliga o cluster e mantém tudo. `make k8s-delete` apaga o namespace, inclusive os dados do MySQL.

A explicação de cada recurso, a referência de comandos `kubectl` e a solução dos problemas mais comuns (HPA em `<unknown>`, pod que não fica pronto, migration com erro) estão em [`docs/KUBERNETES.md`](docs/KUBERNETES.md).
