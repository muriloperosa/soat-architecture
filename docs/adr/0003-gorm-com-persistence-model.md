# 0003. GORM sem vazamento: Persistence Model e Mapper

## Status
Aceita

## Contexto
GORM exige tags `gorm:"..."` e traz comportamentos (lazy loading, hooks, `ErrRecordNotFound`) que não devem contaminar o domínio. Se a entidade de domínio carregar tags GORM, o domínio passa a depender de infraestrutura, o que viola a regra de dependência (`infrastructure` importa `application`/`domain`, nunca o inverso).

## Decisão
Cada agregado persistido ganha, na camada de infraestrutura (`internal/infrastructure/persistence/mysql/<agregado>/`), três arquivos:
- `model.go`: struct de persistência com tags `gorm:"..."`, sem lógica de negócio.
- `mapper.go`: converte entidade em model e vice versa. Usa um construtor de reconstituição (não gera ID novo, não reseta status) para reidratar entidades vindas do banco.
- `repository.go`: implementa a interface de domínio, executa queries e traduz erros na fronteira (`gorm.ErrRecordNotFound` vira erro de domínio, por exemplo `ErrNaoEncontrada`).

## Consequências
Positivas: entidade de domínio permanece pura e testável sem banco. Troca de ORM ou driver não afeta `domain`/`application`. Erros de infraestrutura não vazam para camadas superiores.

Negativas: um arquivo a mais por agregado (mapper) e alguma duplicação estrutural entre entidade e model. Custo aceito em troca do isolamento do domínio.
