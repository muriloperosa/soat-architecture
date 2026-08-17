# 0005. Convenção de nomes: scaffolding técnico em inglês, negócio em português

## Status
Aceita

## Contexto
O projeto mistura dois vocabulários: termos técnicos de arquitetura em camadas (padrão da indústria, majoritariamente em inglês na comunidade Go) e a linguagem ubíqua do domínio de oficina mecânica, que é em português. Nomear tudo em inglês afasta o código dos termos que o time de negócio usa; nomear tudo em português força tradução forçada de padrões consagrados como repository, mapper e model.

## Decisão
Regra híbrida por natureza do nome, não por camada:
- Scaffolding técnico em inglês: nomes de camada e infraestrutura (`domain`, `application`, `infrastructure`, `persistence`) e artefatos de padrão (`repository.go`, `mapper.go`, `model.go`, `router.go`, `jwt.go`, `config.go`, `connection.go`).
- Negócio em português, seguindo a linguagem ubíqua: pacotes de agregado (`ordemservico`, `cliente`) e arquivos de conceito (`ordem_servico.go`, `status.go`, `documento.go`). `shared/` fica em inglês por ser agrupamento técnico, mas os Value Objects dentro dele têm nome de negócio em português.
- Pasta namespaceada por agregado usa nome de arquivo genérico (`model.go`, `repository.go`, `handler.go`, `dto.go`, `mapper.go`, `routes.go`), dentro da pasta do agregado o pacote já desambigua, não precisa prefixo. `internal/infrastructure/http/` segue essa regra: cada domínio ganha pasta própria (`http/auth/`, `http/health/`, `http/usuario/`, `http/<dominio>/`), com nomes sem prefixo (`interno_handler.go`/`cliente_handler.go` continuam distintos por papel, não por domínio, já que os dois vivem dentro de `http/auth/`). `http/httperror/` e `http/httprequest/` são as pastas que não são domínio: concentram, respectivamente, `ErrorResponse`/`Respond*Error` e helpers de request HTTP (`ParseUintParam`) usados por handlers de qualquer domínio; ficam fora da raiz e de qualquer domínio de propósito, porque são pacote-folha (só importam `domain/shared` ou nada do projeto) e isso evita ciclo de import entre `router.go` (raiz, importa cada domínio pra registrar rota) e os domínios. `http/middleware/` segue o mesmo raciocínio pra `SubjectID` (lê o usuário logado do JWT): mora junto de `ClaimsContextKey`/`AuthenticationMiddleware`, que já são a fonte desse dado, em vez de duplicado em cada domínio que precisa de self-service. Só arquivos que não pertencem a nenhum domínio específico, como `router.go`, ficam soltos na raiz de `http/`.
- Cuidado com ciclo de import ao namespacear por domínio: pacote de domínio nunca importa a raiz `http` de volta. Qualquer coisa usada tanto pela raiz quanto pelos domínios (ou só pelos domínios) vira pacote-folha próprio, como `httperror`, nunca fica na raiz.
- Pacotes, isto é o nome da pasta, em ASCII minúsculo, sem acento e sem underscore. Arquivos podem usar underscore.
- `cmd/api/main.go` é exceção deliberada: nome em inglês, padrão idiomático de composition root em Go. `package main` e `func main()` são fixos da linguagem de qualquer forma. `cmd/create-user/main.go` segue a mesma exceção — nome de comando (verbo-substantivo, estilo CLI), não de conceito de negócio.

## Consequências
Positivas: quem lê o código de domínio reconhece os termos que aparecem em reunião com o time de negócio. Quem lê a infraestrutura reconhece os padrões usuais da comunidade Go.

Negativas: a regra exige julgamento em casos de borda (por exemplo `shared/`) e precisa ser documentada, já que não é auto evidente para quem não conhece a convenção. Revisões de código precisam checar consistência de nomes.
