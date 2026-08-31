# 0001. Monólito em camadas com DDD tático

## Status
Aceita

## Contexto
O enunciado do Tech Challenge Fase 1 permite explicitamente "monolito em camadas". O sistema tem poucos bounded contexts (ordem de serviço, cliente, veículo, serviço, peça) e prazo de MVP. Uma arquitetura de microsserviços ou hexagonal completa seria overengineering nesse estágio.

## Decisão
Adotar arquitetura em camadas (`domain` -> `application` -> `infrastructure`) com núcleo tático de DDD: entidades, Value Objects auto validáveis, agregados, repositórios como interface. Simplificações conscientes do MVP:
- Sem Unit of Work completo. Operações multiagregado que exigem atomicidade usam a abstração `TransactionRunner`; a implementação MySQL injeta a `tx` do GORM no `context.Context`, permitindo que os repositórios participem da mesma transação sem vazar GORM para aplicação/domínio.
- Sem barramento de eventos de domínio. Políticas são orquestração direta dentro do use case.

## Consequências
Positivas: onboarding rápido, menos infraestrutura acidental, domínio ainda testável e isolado de Gin e GORM.

Negativas: não escala para múltiplos serviços ou times sem refatoração. Se o número de agregados crescer muito, o `TransactionRunner` baseado em contexto e a ausência de event bus podem virar dívida técnica deliberada, a ser revisitada em uma ADR futura.
