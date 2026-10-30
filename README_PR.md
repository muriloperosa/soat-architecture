# PR: Prioridade e filtros na listagem de ordens de serviço

## Resumo

A listagem padrão de ordens de serviço segue a prioridade operacional definida para a Fase 2:

1. `EM_EXECUCAO`
2. `APROVADA`
3. `AGUARDANDO_APROVACAO` e `REJEITADA` (mesma prioridade)
4. `EM_DIAGNOSTICO`
5. `RECEBIDA`

Ordens mais antigas aparecem primeiro dentro de cada prioridade, usando `data_cadastro ASC`; `id` é o critério final de desempate. `direction=DESC` inverte a ordem e os critérios de desempate. `order=status` mantém a prioridade de negócio; outros campos de `order` aplicam a ordenação solicitada.

Sem filtro explícito de status, `FINALIZADA` e `ENTREGUE` ficam fora da fila e da contagem total. Quando a requisição informa status, a consulta respeita o filtro, permitindo consultar ordens concluídas e o histórico do cliente.

## Implementação

- A prioridade dos status foi centralizada no domínio, agrupando `AGUARDANDO_APROVACAO` e `REJEITADA`.
- O query builder aceita expressões compostas de ordenação com valores SQL parametrizados.
- O repositório aplica a exclusão de status finais somente na ausência de filtro de status.
- A descrição do endpoint `GET /v1/ordens-servico` e os artefatos Swagger foram atualizados.
- Os testes unitários e de integração cobrem a ordem, a direção descendente e a consulta explícita de status concluídos.

## Validação

```bash
go test ./internal/domain/ordemservico ./internal/infrastructure/persistence/mysql/query ./internal/infrastructure/persistence/mysql/ordemservico
go test -tags=integration ./test/integration -run TestConsultaOrdemServico_ListagemOrdenaPorStatusEExcluiFinalizadas -count=1
```
