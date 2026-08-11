# 0004. Gin como framework web

## Status
Aceita

## Contexto
É necessário um framework HTTP para Go com maturidade, bom volume de material de referência (Swagger, JWT, middlewares) e baixo overhead de aprendizado para o time. Echo seria uma alternativa igualmente válida.

## Decisão
Usar Gin (`github.com/gin-gonic/gin`). Regra inegociável: `gin.Context` vive só em `internal/infrastructure/http`, nunca entra em `application` nem em `domain`. Handlers ficam magros: `bind`, `toInput()`, use case, `toResponse()`, resposta. Cada agregado tem DTOs e mappers de transporte próprios.

## Consequências
Positivas: ecossistema maduro, fácil integração com `swaggo` para Swagger e middlewares de JWT. Handlers finos mantêm a aplicação agnóstica de transporte.

Negativas: troca de framework web exigiria reescrever toda a camada `http`, mas não afeta `domain`/`application`. O isolamento é o objetivo desta decisão.
