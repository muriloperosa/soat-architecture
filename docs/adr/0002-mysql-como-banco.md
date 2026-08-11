# 0002. MySQL 8 como banco de dados

## Status
Aceita

## Contexto
O domínio é relacional e exige consistência transacional forte. Gerar orçamento e baixar estoque de peça, por exemplo, precisam ser atômicos. É necessário um banco com suporte maduro em Go, fácil de rodar em `docker compose` e comum em ambientes de PME, contexto do Tech Challenge.

## Decisão
Usar MySQL 8, engine InnoDB. MySQL 8 oferece CTEs, window functions e CHECK constraints, cobrindo as necessidades relacionais do domínio sem exigir um banco mais sofisticado. O driver Go `gorm.io/driver/mysql` é maduro e amplamente documentado.

## Consequências
Positivas: ACID via InnoDB garante atomicidade em operações multi tabela, setup trivial em Docker, grande base de conhecimento disponível.

Negativas: menos recursos avançados que Postgres, por exemplo tipos JSON mais limitados e sem `RETURNING` nativo em todas as versões. Schema é versionado manualmente em `migrations/*.sql`; `AutoMigrate` do GORM é evitado de propósito para manter controle explícito de `up`/`down`.
