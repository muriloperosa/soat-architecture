# 0005. Convenção de nomes: scaffolding técnico em inglês, negócio em português

## Status
Aceita

## Contexto
O projeto mistura dois vocabulários: termos técnicos de arquitetura em camadas (padrão da indústria, majoritariamente em inglês na comunidade Go) e a linguagem ubíqua do domínio de oficina mecânica, que é em português. Nomear tudo em inglês afasta o código dos termos que o time de negócio usa; nomear tudo em português força tradução forçada de padrões consagrados como repository, mapper e model.

## Decisão
Regra híbrida por natureza do nome, não por camada:
- Scaffolding técnico em inglês: nomes de camada e infraestrutura (`domain`, `application`, `infrastructure`, `persistence`) e artefatos de padrão (`repository.go`, `mapper.go`, `model.go`, `router.go`, `jwt.go`, `config.go`, `connection.go`).
- Negócio em português, seguindo a linguagem ubíqua: pacotes de agregado (`ordemservico`, `cliente`) e arquivos de conceito (`ordem_servico.go`, `status.go`, `documento.go`). `shared/` fica em inglês por ser agrupamento técnico, mas os Value Objects dentro dele têm nome de negócio em português.
- Pasta namespaceada por agregado usa nome de arquivo genérico (`model.go`, `repository.go`). Pasta plana usa prefixo por entidade, por exemplo `http/ordem_servico_handler.go`, para desambiguar.
- Pacotes, isto é o nome da pasta, em ASCII minúsculo, sem acento e sem underscore. Arquivos podem usar underscore.
- `cmd/api/main.go` é exceção deliberada: nome em inglês, padrão idiomático de composition root em Go. `package main` e `func main()` são fixos da linguagem de qualquer forma.

## Consequências
Positivas: quem lê o código de domínio reconhece os termos que aparecem em reunião com o time de negócio. Quem lê a infraestrutura reconhece os padrões usuais da comunidade Go.

Negativas: a regra exige julgamento em casos de borda (por exemplo `shared/`) e precisa ser documentada, já que não é auto evidente para quem não conhece a convenção. Revisões de código precisam checar consistência de nomes.
