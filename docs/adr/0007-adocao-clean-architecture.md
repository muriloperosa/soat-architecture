# 0007. Adoção da Clean Architecture

## Status
Aceita. Substitui a [0001](0001-monolito-em-camadas.md).

## Contexto
A [ADR 0001](0001-monolito-em-camadas.md) registrou, na Fase 1, um monólito em camadas com DDD tático e afirmou que uma arquitetura hexagonal completa seria exagero para o MVP. O enunciado da Fase 2 passou a exigir explicitamente Clean Architecture ou Arquitetura Hexagonal, com separação adequada de camadas e dependências.

Revendo o código, a estrutura já era Clean Architecture na prática, só não estava documentada como tal:

- `internal/domain` não importa nada de `application` nem de `infrastructure`.
- Os repositórios e demais dependências externas do domínio são interfaces declaradas no próprio domínio (`ClienteRepository`, `OrdemServicoRepository`, `JWTProvider`, `EmailSender`, `TransactionRunner` e outras), implementadas em `internal/infrastructure`.
- Cada caso de uso é um struct em `internal/application/<agregado>` com DTOs de entrada e saída próprios, sem conhecer HTTP nem GORM.
- Os models do GORM e os DTOs HTTP ficam na infraestrutura, com mappers explícitos nas fronteiras ([ADR 0003](0003-gorm-com-persistence-model.md)).
- `internal/infrastructure/wiring` é o único lugar que conhece todas as implementações e monta o grafo de dependências.

A ADR 0001 continuava como "Aceita" e contradizia tanto o enunciado quanto o código. Ao formalizar a decisão, um teste automático da regra de dependência encontrou três pontos fora dela:

1. `domain/auth.AppClaims` embutia `jwt.RegisteredClaims`, acoplando o domínio à biblioteca de JWT.
2. O login (`application/auth`) comparava a senha chamando o bcrypt direto, em vez de usar o `SenhaHash` do domínio.
3. O value object `shared.SenhaHash` usa `golang.org/x/crypto/bcrypt`.

## Decisão
Adotar Clean Architecture como estilo arquitetural da aplicação. O sistema continua sendo um único deployável; a Clean Architecture organiza o código por dentro e não muda a forma de implantar.

As camadas e o que fica em cada uma:

| Camada da Clean Architecture | Onde está | Conteúdo |
|---|---|---|
| Entities (regras de negócio da empresa) | `internal/domain` | Entidades, agregados, value objects, erros de domínio e as interfaces das dependências externas que o domínio usa (repositórios, `JWTProvider`, `EmailSender`, `TransactionRunner`) |
| Use Cases (regras de negócio da aplicação) | `internal/application` | Um struct por caso de uso, com método `Executar` e DTOs de entrada e saída próprios |
| Interface Adapters | `internal/infrastructure/http`, `internal/infrastructure/persistence` | Handlers, DTOs e mappers HTTP, middlewares; models, mappers e repositórios GORM |
| Frameworks e Drivers | `internal/infrastructure/{auth,email,config,database,migration}`, `cmd/` | Gin, GORM e driver MySQL, biblioteca de JWT, golang-migrate, leitura de variáveis de ambiente, `main` |
| Composition root | `internal/infrastructure/wiring` | Instancia adaptadores e casos de uso e injeta as dependências |

A regra de dependência aponta só para dentro:

- `domain` depende apenas da biblioteca padrão do Go e do próprio `domain`.
- `application` depende apenas da biblioteca padrão, de `domain` e de `application`.
- `infrastructure` e `cmd` podem depender de qualquer camada.

Os três desvios encontrados foram tratados assim:

1. `AppClaims` virou uma struct pura no domínio (`Subject`, `Tipo`, `Papel`, `Jti`). O payload JWT, com os claims registrados e a expiração, passou para um tipo privado no adaptador `infrastructure/auth`, que converte para `AppClaims` ao validar o token. O formato do token emitido não mudou; um teste fixa os campos `sub`, `tipo`, `papel`, `jti`, `iat` e `exp`.
2. O login passou a usar `SenhaHash.Confere`, e `application` não importa mais o bcrypt.
3. O bcrypt no `SenhaHash` foi mantido como exceção explícita. A regra "senha nunca é guardada em texto plano" é do domínio, e o bcrypt é um algoritmo puro, sem I/O nem estado, mantido pelo time do Go. Tirá-lo do domínio exigiria injetar um serviço de hash nos construtores de `Usuario` e `Cliente` e em todos os seus testes, sem ganho real de isolamento. A exceção é a única permitida e está listada em `excecoesDoDominio` no teste de arquitetura; trocar de algoritmo continua sendo uma mudança num único arquivo.

A regra é verificada por `test/architecture/dependency_rule_test.go`, que lê os imports de todos os arquivos de produção de `internal/domain` e `internal/application` e falha se aparecer qualquer pacote fora do permitido. O teste roda com `go test ./...` e, por isso, também no hook de pre-push.

Duas simplificações em relação ao modelo de referência da Clean Architecture, ambas conscientes:

- **Sem interfaces de entrada (input ports).** Os handlers recebem o caso de uso concreto (`*AbrirOrdemServicoUseCase`) em vez de uma interface. Em Go, interfaces são definidas por quem consome, e cada caso de uso tem uma única implementação; uma interface por caso de uso só acrescentaria arquivos. A dependência continua apontando para dentro (o handler conhece o caso de uso, nunca o contrário), que é o que a regra exige.
- **Sem presenters.** O caso de uso devolve um DTO de saída e o handler o converte na resposta HTTP. O mapper de transporte em `infrastructure/http/<dominio>/mapper.go` cumpre o papel do presenter.

As demais decisões da ADR 0001 continuam valendo: núcleo tático de DDD (value objects auto validáveis, agregados, repositórios como interface), `TransactionRunner` para operações que envolvem mais de um agregado e ausência de barramento de eventos.

## Consequências
Positivas:
- A documentação passa a descrever o que o código faz e atende ao requisito da Fase 2.
- A regra de dependência deixa de depender de revisão manual: um import errado em `domain` ou `application` quebra o build de testes.
- O domínio e os casos de uso continuam testáveis sem banco, sem HTTP e sem biblioteca de JWT, com mocks das interfaces do domínio.
- Trocar Gin, GORM ou a biblioteca de JWT afeta só a camada de infraestrutura.

Negativas:
- Cada dado atravessa três mapeamentos (DTO HTTP, DTO da aplicação e entidade, entidade e model GORM). É mais código do que um CRUD direto, aceito em troca do isolamento ([ADR 0003](0003-gorm-com-persistence-model.md)).
- O teste de arquitetura enxerga só imports. Ele não impede, por exemplo, que um caso de uso receba um `*gorm.DB` como `any`; isso continua dependendo de revisão.
- A exceção do bcrypt precisa ser reavaliada se o domínio passar a precisar de outro algoritmo de hash ou de um provedor externo de identidade.
