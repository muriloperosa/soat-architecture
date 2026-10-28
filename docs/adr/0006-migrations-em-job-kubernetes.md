# 0006. Migrations do banco como Job do Kubernetes com a imagem oficial do golang-migrate

## Status
Aceita

## Contexto
Na Fase 1 o schema é aplicado com `make migrate-up`, que roda `go run ./migrations up` na máquina de quem faz o deploy. Com a aplicação indo para o Kubernetes isso deixa de servir: o banco fica dentro do cluster (Minikube) ou numa rede privada (RDS), a pipeline de CI/CD precisa aplicar o schema sem ninguém rodando comando local, e o HPA sobe e derruba réplicas da API a qualquer momento.

Avaliamos quatro formas de rodar as migrations no cluster:

1. **`initContainer` no Deployment da API**, usando a própria imagem da API com o binário de `migrations/main.go` embutido.
2. **Migrar no boot da API**, com os `.sql` embutidos via `go:embed` e `migrate.Up()` chamado no `main` antes de subir o servidor.
3. **Job separado usando a imagem da API**, com o binário de migration copiado para dentro dela.
4. **Job separado usando a imagem oficial `migrate/migrate`**, recebendo os `.sql` por um ConfigMap.

As opções 1 e 2 têm o mesmo problema: a migration roda em todo pod que nasce. Num pico de carga o HPA pode criar quatro réplicas de uma vez, e as quatro disputam o lock do golang-migrate só para descobrir que não há nada a fazer. Uma migration com erro também impede qualquer pod novo de subir. Os que já estavam no ar continuam atendendo, mas deixam de ter substitutos: o rollout trava, o HPA não consegue escalar e um pod que reinicie não volta. O erro aparece como `Init:Error` ou `CrashLoopBackOff` nos pods da API, misturado com os problemas da própria aplicação. As duas também colocam código e arquivos de schema dentro da imagem que serve tráfego, aumentando o que precisa ser auditado nela.

A opção 3 resolve o "roda uma vez", mas mantém a imagem da API carregando o binário e os `.sql`.

## Decisão
Rodar as migrations num `Job` do Kubernetes (`k8s/migrate/job.yaml`) com a imagem oficial `migrate/migrate`, na mesma versão do `github.com/golang-migrate/migrate/v4` declarada no `go.mod` (v4.19.1). Os arquivos de `migrations/mysql/` chegam ao Job por um ConfigMap `oficina-migrations`, gerado no momento do deploy com `kubectl create configmap --from-file=migrations/mysql`. Não existe cópia versionada em YAML: a pasta `migrations/mysql/` continua sendo a única fonte do schema, e o `make migrate-up` local e os testes de integração seguem usando os mesmos arquivos.

O Job fica numa kustomization própria, fora dos overlays da aplicação, porque o `spec.template` de um Job é imutável: se ele estivesse no mesmo `kubectl apply -k` da API, qualquer mudança nos `.sql` faria o re-apply falhar. O alvo `make k8s-migrate` apaga o Job anterior, recria o ConfigMap, aplica o Job e espera a conclusão. Se a migration falhar, o alvo imprime o log do Job e sai com erro, o que interrompe o `make k8s-deploy` (e, no futuro, o pipeline) antes do restart da API.

A string de conexão é montada pelo kubelet a partir de variáveis de ambiente que vêm do ConfigMap `oficina-api-config` e do Secret `oficina-api-secrets`, os mesmos que a API usa. Os parâmetros do DSN são os de `buildMigrationDSN` em `internal/infrastructure/persistence/mysql/migration_connection.go`.

## Consequências
Positivas:
- As migrations rodam uma vez por deploy, não uma vez por pod. Escalar a API não toca no banco.
- Uma migration quebrada falha o deploy num recurso próprio (`kubectl logs job/oficina-migrate`), sem derrubar as réplicas que já estavam no ar.
- A imagem da API continua com um único binário. O Dockerfile não muda.
- O mesmo Job serve para Minikube e EKS; só muda de onde vêm `DB_HOST` e as credenciais.
- O passo "deploy do banco de dados" do pipeline de CI/CD vira uma chamada a `make k8s-migrate`.

Negativas:
- O ConfigMap e o Secret da aplicação precisam ter nome fixo (`disableNameSuffixHash`), porque o Job é aplicado por outra kustomization e não enxerga o sufixo de hash que o Kustomize geraria. Com isso, alterar um valor de configuração não dispara rollout da API sozinho; o `make k8s-deploy` faz um `kubectl rollout restart` no final para compensar.
- A senha do banco vai dentro de uma URL (`mysql://usuario:senha@tcp(...)`). Caracteres como `@`, `:`, `/`, `?` e `#` quebram o parse, então a senha precisa ser alfanumérica. Os arquivos `secrets.env.example` avisam disso.
- No primeiro deploy os pods da API sobem alguns segundos antes das tabelas existirem, porque o namespace, o MySQL e o Secret são criados pelo mesmo `apply` que cria o Deployment. O restart ao final do `k8s-deploy` resolve. Nos deploys seguintes as migrations precisam ser compatíveis com a versão anterior da API (primeiro adiciona coluna ou tabela; a remoção fica para um deploy posterior), já que réplicas antigas continuam atendendo enquanto o Job roda.
- O ConfigMap é aplicado com `kubectl apply` no modo client-side, que guarda uma cópia do objeto inteiro na anotação `last-applied-configuration`. Anotações têm limite de 256 KiB, e esse é o teto prático para a soma dos `.sql` (o limite do próprio ConfigMap é 1 MiB). Hoje `migrations/mysql/` soma cerca de 15 KB. Se um dia passar de uns 200 KB, trocar para `kubectl apply --server-side`, que não grava essa anotação.
- Se uma migration falhar no meio, o golang-migrate marca a versão como `dirty` e recusa novas execuções até alguém corrigir o banco e rodar `force` manualmente. O procedimento está em `docs/KUBERNETES.md`, na seção de problemas comuns.
