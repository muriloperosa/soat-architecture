# Sistema de Oficina Mecânica: Tech Challenge Fase 1

API de gestão de ordens de serviço de uma oficina mecânica, cobrindo o fluxo completo desde o cadastro do cliente e do veículo até a execução do serviço, passando por diagnóstico, orçamento, aprovação, reserva de peças e entrega. Back-end em Go, com Gin, GORM e MySQL 8.

## Overview de Arquitetura

O projeto segue um monólito em camadas com núcleo tático de DDD: domínio, aplicação e infraestrutura são pacotes separados, com regra de dependência em uma única direção (infraestrutura conhece aplicação e domínio, aplicação conhece só domínio, domínio não conhece nada do projeto). Cada agregado (ordem de serviço, cliente, veículo, peça, usuário) carrega suas próprias regras de negócio em Value Objects auto validáveis, como `Documento`, `Placa` e `Status`, que nascem válidos ou não nascem.

A persistência usa GORM, mas de forma isolada: os models do GORM nunca vazam para a camada de domínio, existe um mapeamento explícito entre entidade e model na fronteira do repositório. Autenticação é feita via JWT, com um middleware que reavalida se o usuário segue ativo a cada requisição, não só no login, e um segundo middleware que autoriza por tipo e papel de usuário. Situações que exigem consistência entre agregados diferentes, como reservar peça e verificar estoque, passam por um `TransactionRunner` que garante atomicidade e usa `SELECT ... FOR UPDATE` para serializar concorrência.

O versionamento do schema do banco é feito com `golang-migrate`, migrations em `migrations/mysql/`, aplicadas via `make migrate-up`. As decisões estruturais do projeto (monólito em camadas, MySQL, GORM sem vazamento, framework Gin, convenção de nomes) estão registradas como ADRs em [`docs/adr/`](docs/adr/), e o detalhamento das convenções internas está em [`docs/ARQUITETURA.md`](docs/ARQUITETURA.md).

## Documentação

- [Guia de desenvolvimento](docs/DEVELOPMENT_GUIDE.md): pacote completo das informações necessárias para o desenvolvimento do projeto, serve como guia para os desenvolvedores
- [Arquitetura](docs/ARQUITETURA.md): convenções internas de camadas, mapeamento e injeção de dependência
- [ADRs](docs/adr/): decisões estruturais registradas e justificadas
- [Modelo de domínio](docs/DOMAIN_MODEL.md): agregados, entidades e Value Objects
- [Segurança](docs/SECURITY.md): ferramentas de SCA, SAST e DAST usadas e a justificativa de cada uma
- [Swagger](docs/swagger/): documentação da API gerada a partir dos handlers (UI disponível em `/swagger/index.html` com a aplicação no ar)
- [Coleção Postman](docs/postman/): coleção pronta para testar os endpoints da API

## Qualidade

O projeto mantém uma meta de cobertura de 80% ou mais nos domínios críticos, com testes unitários isolados por camada via mocks (`mockery`) e testes de integração que sobem um MySQL real com `testcontainers`, aplicam as migrations de produção e batem no router completo da aplicação, sem nenhum tipo de mock. `make lint` roda `go vet` e as regras estáticas do projeto antes de qualquer commit ser aceito.

A análise de vulnerabilidades cobre três frentes independentes: `govulncheck` para dependências conhecidas (SCA), `gosec` para padrões de risco no código fonte (SAST) e OWASP ZAP para ataques simulados contra a API em execução (DAST), este último rodando autenticado uma vez por papel de usuário para provar que o controle de acesso bloqueia de fato quem não tem permissão. SonarQube fica disponível localmente para análise estática complementar. Todos os detalhes e a justificativa de cada ferramenta estão em [`docs/SECURITY.md`](docs/SECURITY.md).

Um hook de `pre-push` local, instalado via `make hooks-install`, roda mocks, lint, testes e geração do Swagger antes de liberar qualquer push, evitando que código desatualizado ou quebrado chegue ao repositório remoto.

## Executar Localmente

Instruções completas de configuração, execução via Docker ou em modo dev local, testes e demais comandos estão no [Guia de Desenvolvimento](docs/DEVELOPMENT_GUIDE.md).
