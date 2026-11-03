# Sistema de Oficina Mecânica: Tech Challenge Fase 2

[![CI](https://github.com/muriloperosa/soat-architecture/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/muriloperosa/soat-architecture/actions/workflows/ci.yml)
[![CD](https://github.com/muriloperosa/soat-architecture/actions/workflows/cd.yml/badge.svg)](https://github.com/muriloperosa/soat-architecture/actions/workflows/cd.yml)

API de gestão de ordens de serviço de uma oficina mecânica, cobrindo o fluxo completo desde o cadastro do cliente e do veículo até a execução do serviço, passando por diagnóstico, orçamento, aprovação, reserva de peças e entrega. Back-end em Go, com Gin, GORM e MySQL 8.

Na Fase 2 a aplicação passou a rodar em Kubernetes, no Minikube para desenvolvimento e no Amazon EKS na nuvem, com o deploy feito pelo GitHub Actions.

> [!IMPORTANT]
> **Vídeo demonstrativo:** [youtube.com/watch?v=VIDEO_ID](https://www.youtube.com/watch?v=VIDEO_ID) (até 15 minutos).

## Objetivos da Fase 2

O sistema da Fase 1 já atendia a oficina, mas com mais demanda e a abertura de novas unidades ele precisa aguentar horários de pico e não pode depender de deploy manual. Nesta fase trabalhamos em:

- infraestrutura que escala, para diminuir o risco de o sistema ficar fora do ar;
- provisionamento e deploy automatizados;
- código mais organizado e com testes, para facilitar a evolução;
- escala automática da API quando o volume de ordens de serviço aumenta.

Onde encontrar cada requisito da fase no repositório:

| Requisito | Onde está |
| --- | --- |
| Clean Code, Clean Architecture e testes automatizados | Camadas em `internal/`, regra de dependência verificada em `test/architecture/`, testes unitários e de integração ([Qualidade](#qualidade)) |
| Abertura, consulta de status e listagem priorizada de OS | `/v1/ordens-servico` ([APIs](#apis)) |
| Aprovação ou recusa do orçamento pelo cliente | Rotas autenticadas `/v1/ordens-servico/{id}/orcamento/aprovar` e `/rejeitar`; links por e-mail em `GET /v1/orcamentos/aprovacao` (confirmação) e `POST /v1/orcamentos/aprovacao/decisao` (decisão), que reutilizam os mesmos casos de uso |
| Aviso de mudança de status por e-mail | Provedor-alvo: Amazon SES. No AWS Academy o envio usa SMTP com STARTTLS, porque o Learner Lab não libera o SES ([Componentes](#componentes-da-aplicação)) |
| Dockerfile e docker-compose | [`Dockerfile`](Dockerfile), [`compose.yml`](compose.yml) |
| Deployments, Services, ConfigMaps, Secrets e HPA | [`k8s/`](k8s/) |
| Terraform do cluster Kubernetes | [`infra/`](infra/) |
| Banco de dados na nuvem | RDS MySQL, criado seguindo [`docs/AWS_SETUP.md`](docs/AWS_SETUP.md) |
| Pipeline de CI/CD | [`.github/workflows/`](.github/workflows/) |

## Stack

| Camada | Tecnologia |
| --- | --- |
| Linguagem e HTTP | Go 1.26, Gin |
| Persistência | MySQL 8, GORM, `golang-migrate` |
| Autenticação | JWT HS256 com access e refresh token |
| E-mail | Amazon SES como provedor-alvo; SMTP com STARTTLS no AWS Academy, onde o SES não é liberado; provedor `log` para desenvolvimento |
| Documentação da API | Swagger gerado com `swaggo`, collections Postman |
| Contêiner | Docker multi-stage (`golang:1.26-alpine` → `alpine`, usuário sem privilégio) e Docker Compose |
| Orquestração | Kubernetes com Kustomize: Minikube e Amazon EKS |
| Infraestrutura como código | Terraform 1.11+ com o provider AWS 5.x e state no S3 |
| CI/CD | GitHub Actions, com Terraform no deploy e state no S3 (AWS Academy Learner Lab) |
| Registry | Amazon ECR, com tags imutáveis |
| Testes e análise | `testify`, `mockery`, `testcontainers`, k6, `govulncheck`, `gosec`, OWASP ZAP, SonarQube |

## Arquitetura

### Componentes da aplicação

O projeto segue Clean Architecture com núcleo tático de DDD, num único deployável.

```mermaid
flowchart TB
    req(["Requisição HTTP"]) --> mw

    subgraph infra["Infraestrutura · internal/infrastructure"]
        direction LR
        mw["Middlewares<br/>autenticação JWT<br/>autorização por papel"]
        handlers["Handlers<br/>DTOs e mappers"]
        persist["Repositórios GORM<br/>TransactionRunner"]
        email["EmailSender<br/>log, SMTP ou SES"]
        mw --> handlers
    end

    subgraph app["Aplicação · internal/application"]
        uc["Casos de uso"]
    end

    subgraph dom["Domínio · internal/domain"]
        direction LR
        ent["Agregados e<br/>Value Objects"]
        ports["Interfaces<br/>repositórios e serviços"]
    end

    handlers --> uc
    uc --> ent
    uc --> ports
    persist -. implementa .-> ports
    email -. implementa .-> ports
    persist --> db[("MySQL 8")]
    email -->|alvo| ses[["Amazon SES"]]
    email -.->|AWS Academy| smtp[["Servidor SMTP<br/>STARTTLS"]]
```

Domínio (entidades e value objects), aplicação (casos de uso) e infraestrutura (adaptadores HTTP, de persistência e de e-mail, frameworks) são pacotes separados, e as dependências apontam só para dentro: infraestrutura conhece aplicação e domínio, aplicação conhece só domínio, e domínio não conhece nada do projeto. Os repositórios e serviços externos que o domínio usa são interfaces declaradas nele e implementadas na infraestrutura. Essa regra é verificada por um teste automático (`test/architecture/`). Cada agregado (ordem de serviço, cliente, veículo, peça, usuário) carrega suas próprias regras de negócio em Value Objects auto validáveis, como `Documento`, `Placa` e `Status`, que nascem válidos ou não nascem.

A persistência usa GORM, mas de forma isolada: os models do GORM nunca vazam para a camada de domínio, existe um mapeamento explícito entre entidade e model na fronteira do repositório. Autenticação é feita via JWT, com um middleware que reavalida se o usuário segue ativo a cada requisição, não só no login, e um segundo middleware que autoriza por tipo e papel de usuário. Situações que exigem consistência entre agregados diferentes, como reservar peça e verificar estoque, passam por um `TransactionRunner` que garante atomicidade e usa `SELECT ... FOR UPDATE` para serializar concorrência.

Cada mudança de status da ordem de serviço dispara um e-mail ao cliente depois que a transição já foi persistida. Se o envio falhar, o erro fica no log e o status não é desfeito. No envio do orçamento vai um segundo e-mail, com serviços, peças, quantidades e valores. Localmente o provedor padrão é `log` (`EMAIL_PROVIDER`), que só registra a mensagem.

O provedor-alvo é o **Amazon SES** (`EMAIL_PROVIDER=ses`), e o adapter está pronto e testado. Numa conta AWS comum, ele usa a role dos nós e exige só um remetente verificado em `SES_FROM_ADDRESS`. O AWS Academy Learner Lab, onde o projeto roda, nega `ses:VerifyEmailIdentity`, e sem identidade verificada o SES não envia nada. Por isso, no Academy, o envio usa o provedor `smtp`: SMTP autenticado com STARTTLS (Gmail: `smtp.gmail.com:587`, com senha de app). Fora do Academy, a aplicação volta ao SES trocando `EMAIL_PROVIDER` e configurando o remetente, sem mudança de código; o que muda no pipeline está na [ADR 0010](docs/adr/0010-smtp-no-lugar-do-ses-no-learner-lab.md). No SMTP, configure `SMTP_HOST`, `SMTP_PORT`, `SMTP_USERNAME`, `SMTP_PASSWORD` e `SMTP_FROM_ADDRESS`; não versione a senha de app. No EKS, o Terraform grava a senha no Secrets Manager com um argumento write-only, sem persistir seu valor no state ou no plano.

Quando a OS entra em `AGUARDANDO_APROVACAO`, o e-mail do orçamento inclui links temporários para aceitar ou recusar. No deploy de produção, o Terraform gera `APPROVAL_TOKEN_SECRET` e o guarda no Secrets Manager junto com `JWT_SECRET`; o workflow lê o segredo, monta a URL dos links com o hostname do LoadBalancer e injeta ambos no Secret Kubernetes. Para outro endereço, como um domínio HTTPS, defina a variável `PUBLIC_BASE_URL` no environment `aws-deploy`. O segredo de aprovação é distinto e permanece estável entre deploys; rotacioná-lo invalida links ainda válidos. A validade padrão é de 24 horas (`APPROVAL_TOKEN_TTL_HOURS`). O link abre uma página de confirmação; somente o `POST` de decisão altera o orçamento. Os endpoints validam o token e reutilizam os mesmos casos de uso das rotas autenticadas. O `GET` não altera o estado, para evitar decisões causadas por scanners de e-mail. O fluxo usa a própria API e não exige webhook externo. O LoadBalancer atual expõe HTTP; para uso com clientes reais, publique a API via HTTPS antes de enviar os links. Consulte [AWS_SETUP.md](docs/AWS_SETUP.md#4-links-de-aprovação-do-orçamento).

O versionamento do schema do banco é feito com `golang-migrate`, migrations em `migrations/mysql/`. As decisões estruturais do projeto estão registradas como ADRs em [`docs/adr/`](docs/adr/), e o detalhamento das convenções internas está em [`docs/ARQUITETURA.md`](docs/ARQUITETURA.md).

### Ciclo de vida da ordem de serviço

As transições de status são regras do domínio (`StatusOrdemServico.PermiteTransicaoPara`), e qualquer outra é recusada pela API.

```mermaid
stateDiagram-v2
    [*] --> RECEBIDA: abrir OS
    RECEBIDA --> EM_DIAGNOSTICO: iniciar diagnóstico
    EM_DIAGNOSTICO --> AGUARDANDO_APROVACAO: enviar orçamento
    AGUARDANDO_APROVACAO --> APROVADA: cliente aprova
    AGUARDANDO_APROVACAO --> REJEITADA: cliente recusa
    REJEITADA --> AGUARDANDO_APROVACAO: reenviar orçamento
    APROVADA --> AGUARDANDO_APROVACAO: reenviar orçamento
    APROVADA --> EM_EXECUCAO: iniciar execução
    EM_EXECUCAO --> FINALIZADA: finalizar
    FINALIZADA --> ENTREGUE: entregar
    ENTREGUE --> [*]
```

Quando o cliente aprova o orçamento, as peças ficam reservadas, e a baixa no estoque só acontece na finalização. A listagem de ordens de serviço segue a ordem de prioridade da oficina: em execução, aprovadas, aguardando aprovação ou rejeitadas, em diagnóstico e recebidas, das mais antigas para as mais novas. As finalizadas e entregues saem da listagem, mas continuam disponíveis na consulta por id ou por número.

### Infraestrutura provisionada

Na AWS, a aplicação fica organizada assim. Em verde está o que o Terraform cria, em azul o que os manifestos de [`k8s/`](k8s/) criam a cada deploy e em cinza tracejado o Amazon SES, provedor-alvo de e-mail que o Learner Lab não libera. No AWS Academy, o e-mail sai por SMTP (Gmail).

```mermaid
flowchart TB
    usuario(["Usuários da API"]) --> lb

    subgraph aws["AWS · us-east-1"]
        lb["Load Balancer<br/>criado pelo Service"]
        ecr[("ECR<br/>oficina-api")]
        sm["Secrets Manager<br/>oficina/api e oficina/smtp"]
        ses[["Amazon SES<br/>provedor-alvo<br/>bloqueado no Learner Lab"]]

        subgraph vpc["VPC default · subnets públicas"]
            rds[("RDS MySQL")]

            subgraph eks["Cluster EKS oficina · Terraform"]
                nodes["Node group · Terraform<br/>t3.small, 1 a 2 nós"]

                subgraph ns["namespace oficina · manifestos k8s/"]
                    svc["Service<br/>LoadBalancer"]
                    pods["Pods da API<br/>Deployment"]
                    hpa["HPA<br/>2 a 6 réplicas"]
                    job["Job<br/>oficina-migrate"]
                    svc --> pods
                    hpa -->|escala| pods
                end
            end
        end
    end

    lb --> svc
    nodes -.->|executa| pods
    pods --> rds
    job -->|migrate up| rds
    pods -.->|fora do Academy| ses
    pods -->|AWS Academy| smtp[["Gmail SMTP<br/>STARTTLS · porta 587"]]
    ecr -.->|imagem| pods
    sm -.->|segredos no deploy| ns

    classDef terraform fill:#dcfce7,stroke:#16a34a,color:#14532d
    classDef k8s fill:#dbeafe,stroke:#2563eb,color:#1e3a8a
    classDef target fill:#f3f4f6,stroke:#6b7280,color:#374151,stroke-dasharray:4 3
    class nodes,rds,sm,ecr terraform
    class ses target
    class svc,pods,hpa,job k8s
    style eks fill:#f0fdf4,stroke:#16a34a,color:#14532d
```

| Recurso | Como é criado |
| --- | --- |
| Cluster EKS, node group, RDS MySQL, repositório ECR e secrets `oficina/api` e `oficina/smtp` | Terraform em [`infra/`](infra/), na VPC default e nas subnets públicas que já existem na conta |
| Deployment, Service, HPA, ConfigMap, Secret e Job de migration | Manifestos de [`k8s/`](k8s/), aplicados pelo deploy |
| Load Balancer | Criado pela AWS a partir do Service `LoadBalancer` do overlay EKS |
| Credenciais da sessão do Learner Lab no environment `aws-deploy` do GitHub | A cada sessão do lab, seguindo [`docs/AWS_SETUP.md`](docs/AWS_SETUP.md) |

No Minikube roda o mesmo conjunto, mas com o MySQL dentro do cluster (StatefulSet) e o Service como `NodePort`. Mais detalhes sobre cada recurso em [`docs/KUBERNETES.md`](docs/KUBERNETES.md).

### Fluxo de deploy

```mermaid
flowchart TB
    pr(["Pull request"]) --> ci["CI<br/>lint, testes, cobertura,<br/>SCA, SAST e imagem"]
    ci -->|checks verdes + review| main["Merge na main"]
    main -->|git tag 1.0.1| guard["guard<br/>tag na main e CI verde?"]
    guard --> tf["infra<br/>terraform apply<br/>(state no S3)"]
    tf --> img["image<br/>build e push no ECR"]
    img --> dep["deploy<br/>migrations e rollout no EKS"]
```

A CI ([`ci.yml`](.github/workflows/ci.yml)) roda em todo PR e em todo push na `main`, e não tem acesso à AWS. O deploy ([`cd.yml`](.github/workflows/cd.yml)) só acontece quando alguém cria uma tag de versão, como `1.0.1`, num commit da `main` que passou na CI. O mesmo run aplica o Terraform de [`infra/`](infra/) antes de publicar a imagem, então não há `apply` manual. O workflow usa as credenciais da sessão do AWS Academy Learner Lab, guardadas no environment `aws-deploy`.

Antes de atualizar a API, o deploy roda as migrations num Job ([ADR 0006](docs/adr/0006-migrations-em-job-kubernetes.md)). Se a migration falhar, a versão anterior continua no ar. O passo a passo, as credenciais, o rollback e os custos estão em [`docs/CI_CD.md`](docs/CI_CD.md).

O banco do RDS começa vazio. Depois do primeiro deploy, rode o workflow manual [`seed.yml`](.github/workflows/seed.yml) em **Actions → Seed (EKS) → Run workflow**: escolha a tag implantada em "Use workflow from" e digite `seed` na confirmação. Ele cria os usuários de teste (admin, atendente, mecânico e cliente, senha `Senha@123`) e dados de exemplo. Detalhes em [Popular o banco](docs/CI_CD.md#popular-o-banco-seed).

## APIs

A documentação da API é gerada com Swagger a partir das anotações dos handlers e fica em [`docs/swagger/`](docs/swagger/). Com a aplicação rodando, a UI abre em `/swagger/index.html` e faz as chamadas para o mesmo endereço em que foi aberta, seja localhost, Minikube ou EKS.

Também deixamos três collections do Postman em [`.dev/postman/`](.dev/postman/):

- `soat-architecture.postman_collection.json`, com todos os endpoints;
- `soat-demo-fluxo-completo.postman_collection.json`, que percorre o fluxo da OS do login até a entrega;
- `soat-fase-2.postman_collection.json`, focada nos requisitos de API do PDF da Fase 2, incluindo listagem priorizada com e sem filtro de status. Pode ser executada no Collection Runner; a variável `decisao_orcamento` seleciona aprovação ou rejeição.

A variável `base_url` aponta para o EKS. Para testar localmente, use o valor de `local_url` (`http://localhost:8080`).

## Executar localmente

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

A API fica em `http://localhost:8080` e o Swagger em `http://localhost:8080/swagger/index.html`.

## Provisionar a infraestrutura com Terraform

O Terraform fica em [`infra/`](infra/) e cria o cluster EKS com um node group, o RDS MySQL, o repositório ECR e os secrets `oficina/api` e `oficina/smtp` no Secrets Manager, numa conta do AWS Academy Learner Lab. Cluster e nós usam a `LabRole` do lab. O SES, provedor-alvo de e-mail, não é provisionado porque o Learner Lab nega a verificação de identidade (`ses:VerifyEmailIdentity`). No lugar dele, a pipeline recebe a senha SMTP pelo Environment Secret `SMTP_PASSWORD` no GitHub e grava-a com argumento write-only. O state fica num bucket S3, criado pelo próprio pipeline na primeira execução.

Quem aplica é o CD: cada tag `MAJOR.MINOR.PATCH` na `main` roda o `terraform apply` antes de publicar a imagem e fazer o deploy. Antes da tag, atualize no GitHub as credenciais da sessão do lab ([`docs/AWS_SETUP.md`](docs/AWS_SETUP.md)).

Para rodar à mão (plan, ajuste ou destroy), os pré-requisitos são Terraform 1.11 ou mais novo, AWS CLI, `kubectl` e as credenciais da sessão do lab no profile `oficina-lab`. O plan/apply também precisa receber `TF_VAR_smtp_password` pelo ambiente. Não passe a senha como argumento de linha de comando nem grave a access key no repositório.

**1. Ver o plano ou aplicar.** Na raiz do repositório, usando o mesmo state do pipeline:

```bash
make tf-plan
make tf-apply
```

**2. Apontar o `kubectl` para o cluster:**

```bash
aws eks update-kubeconfig --region us-east-1 --name oficina --profile oficina-lab
```

**3. Destruir no fim da demonstração.** O control plane é cobrado por hora enquanto existir, mesmo com a sessão do lab encerrada:

```bash
make tf-destroy CONFIRM=yes
```

O que é criado, o state no S3 e o roteiro completo de `plan`, `apply` e `destroy` estão em [`docs/TERRAFORM.md`](docs/TERRAFORM.md).

## Deploy em Kubernetes

Os manifestos ficam em [`k8s/`](k8s/), organizados com Kustomize: uma base comum (Deployment, Service, ConfigMap e HPA da API) e overlays para Minikube e EKS. As migrations rodam num Job separado ([ADR 0006](docs/adr/0006-migrations-em-job-kubernetes.md)).

### Minikube: subir e demonstrar o autoscaling

Pré-requisitos: Docker rodando, Minikube, `kubectl`, k6, Go 1.26 e `make`. Versões mínimas e recursos de máquina estão em [`docs/KUBERNETES.md`](docs/KUBERNETES.md#requisitos-da-máquina).

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

### EKS

Pré-requisitos: a conta da AWS preparada seguindo [`docs/AWS_SETUP.md`](docs/AWS_SETUP.md), uma vez por conta, para o repositório no ECR e a role que o GitHub Actions usa no deploy, e a infraestrutura criada pelo [Terraform](#provisionar-a-infraestrutura-com-terraform), com o nó já `Ready`.

**1. Lançar uma versão.** Basta criar uma tag de versão num commit da `main`. Isso dispara o [`cd.yml`](.github/workflows/cd.yml), que confere se a CI do commit passou, publica a imagem no ECR, roda as migrations e atualiza a API.

```bash
git switch main
git pull
git tag 1.0.1
git push origin 1.0.1
```

**2. Acompanhar o deploy** em **Actions → CD (Continuous Deployment)**. No fim do run, o resumo mostra a versão, a imagem e a URL da API. Também dá para pegar a URL direto no cluster:

```bash
kubectl -n oficina get svc oficina-api -o jsonpath='{.status.loadBalancer.ingress[0].hostname}'
```

**3. Conferir** com `curl http://<url>/v1/health/ready`. O Swagger fica em `http://<url>/swagger/index.html`.

O HPA precisa do `metrics-server`, que não vem instalado no EKS. Sem ele, as metas do HPA ficam em `<unknown>`. Rollback, credenciais e o deploy manual com `make k8s-deploy K8S_OVERLAY=eks` estão em [`docs/CI_CD.md`](docs/CI_CD.md#parte-2-deploy) e [`docs/KUBERNETES.md`](docs/KUBERNETES.md#eks).

## Qualidade

O projeto mantém uma meta de cobertura de 80% ou mais nos domínios críticos, com testes unitários isolados por camada via mocks (`mockery`) e testes de integração que sobem um MySQL real com `testcontainers`, aplicam as migrations de produção e batem no router completo da aplicação, sem nenhum tipo de mock. `make lint` roda `go vet` e as regras estáticas do projeto antes de qualquer commit ser aceito.

A análise de vulnerabilidades cobre três frentes independentes: `govulncheck` para dependências conhecidas (SCA), `gosec` para padrões de risco no código fonte (SAST) e OWASP ZAP para ataques simulados contra a API em execução (DAST), este último rodando autenticado uma vez por papel de usuário para provar que o controle de acesso bloqueia de fato quem não tem permissão. SonarQube fica disponível localmente para análise estática complementar. Todos os detalhes e a justificativa de cada ferramenta estão em [`docs/SECURITY.md`](docs/SECURITY.md).

Um hook de `pre-push` local, instalado via `make hooks-install`, roda mocks, lint, testes e geração do Swagger antes de liberar qualquer push, evitando que código desatualizado ou quebrado chegue ao repositório remoto.

## Documentação

- [Guia de desenvolvimento](docs/DEVELOPMENT_GUIDE.md): pacote completo das informações necessárias para o desenvolvimento do projeto, serve como guia para os desenvolvedores
- [Arquitetura](docs/ARQUITETURA.md): convenções internas de camadas, mapeamento e injeção de dependência
- [ADRs](docs/adr/): decisões estruturais registradas e justificadas
- [Modelo de domínio](docs/DOMAIN_MODEL.md): agregados, entidades e Value Objects
- [Banco de dados](docs/database/DATABASE.md): modelo relacional e diagrama
- [Segurança](docs/SECURITY.md): ferramentas de SCA, SAST e DAST usadas e a justificativa de cada uma
- [Kubernetes](docs/KUBERNETES.md): manifestos em `k8s/`, requisitos da máquina, deploy no Minikube, comandos `kubectl` e demonstração do autoscaling
- [Infraestrutura AWS](docs/TERRAFORM.md): Terraform do cluster EKS, do RDS, do ECR e do secret, aplicado pelo deploy por tag
- [Preparação do Learner Lab](docs/AWS_SETUP.md): environment e credenciais da sessão do lab no GitHub
- [CI/CD](docs/CI_CD.md): CI em GitHub Actions (testes, cobertura, segurança) e deploy no EKS por tag, com diagramas de cada etapa, credenciais e rollback
- [Swagger](docs/swagger/): documentação da API gerada a partir dos handlers
- [Collections Postman](.dev/postman/): todos os endpoints, o fluxo completo da demo e o recorte de APIs da Fase 2
