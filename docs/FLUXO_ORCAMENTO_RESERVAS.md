# Fluxo de Orçamento e Reserva de Peças

Este documento descreve o comportamento implementado para orçamento, aprovação do cliente e reserva de peças da Ordem de Serviço.

## Princípios

- `Orcamento` não possui status próprio. A aprovação/rejeição é representada pelo status da `OrdemServico`.
- `ReservaPeca` não possui endpoint público de criação, alteração ou remoção.
- A quantidade reservada é derivada dos `ItemPeca` do orçamento aprovado.
- `reservas_pecas` é a única fonte de verdade para quantidades reservadas; `pecas` mantém apenas estoque físico e estoque mínimo.
- A aprovação da OS e a criação das reservas são atômicas.

## Rotas do orçamento

Rotas internas, restritas a mecânico ou administrador:

```http
POST   /v1/ordens-servico/:id/orcamento
POST   /v1/ordens-servico/:id/orcamento/itens-servico
POST   /v1/ordens-servico/:id/orcamento/itens-peca
DELETE /v1/ordens-servico/:id/orcamento/itens-servico/:itemId
DELETE /v1/ordens-servico/:id/orcamento/itens-peca/:itemId
PATCH  /v1/ordens-servico/:id/orcamento/itens-peca/:itemId/quantidade
PATCH  /v1/ordens-servico/:id/orcamento/enviar-aprovacao
```

Rotas do cliente proprietário da OS:

```http
PATCH /v1/ordens-servico/:id/orcamento/aprovar
PATCH /v1/ordens-servico/:id/orcamento/rejeitar
```

Não existem rotas como:

```http
POST   /v1/reservas-pecas
PUT    /v1/reservas-pecas/:id
DELETE /v1/reservas-pecas/:id
```

## Primeiro envio para aprovação

O orçamento é preparado enquanto a OS está em `EM_DIAGNOSTICO`.

```text
EM_DIAGNOSTICO
  ↓ gerar/editar orçamento
PATCH /orcamento/enviar-aprovacao
  ↓
AGUARDANDO_APROVACAO
  ↓ e-mail para o cliente
```

O envio exige orçamento válido e não vazio.

## Aprovação do cliente

Quando o cliente proprietário aprova, `AprovarOrcamentoUseCase` executa a aprovação e a reserva dentro da mesma transação.

```text
AGUARDANDO_APROVACAO
  ↓ cliente aprova
BEGIN
  ↓
carregar orçamento
  ↓
agrupar ItemPeca por pecaID
  ↓
bloquear peças em ordem crescente de ID
  ↓
ler reservas correntes com FOR UPDATE
  ↓
validar disponibilidade e estoque mínimo
  ↓
criar ReservaPeca para cada peça do orçamento
  ↓
atualizar OS para APROVADA
  ↓
COMMIT
```

Se houver qualquer falha, inclusive estoque insuficiente, ocorre `ROLLBACK`. A OS não permanece parcialmente aprovada e nenhuma reserva parcial deve sobreviver.

### Regra de disponibilidade

Para cada peça:

```text
quantidadeDisponivel = quantidadeEmEstoque - quantidadeJaReservada
```

A nova reserva somente pode ser criada quando:

```text
quantidadeEmEstoque
- quantidadeJaReservada
- quantidadeDoOrcamento
>= estoqueMinimo
```

Quando o orçamento contém mais de um `ItemPeca` para a mesma peça, as quantidades são somadas antes da validação e da criação da reserva, preservando a unicidade de `ordem_servico_id + peca_id`.

## Concorrência

O fluxo usa `SELECT ... FOR UPDATE` para serializar aprovações que disputam a mesma peça.

As peças são bloqueadas em ordem determinística de ID para reduzir risco de deadlock. Depois do lock da peça, as reservas usadas no cálculo são lidas como locking/current read (`FOR UPDATE`), evitando que uma transação reutilize um snapshot anterior e aprove estoque já comprometido por outra transação.

O teste de integração `TestAprovarOrcamento_ConcorrenciaNaoUltrapassaEstoqueDisponivel` valida esse comportamento com aprovações concorrentes reais contra MySQL.

## Rejeição

O cliente pode rejeitar apenas enquanto a OS estiver em `AGUARDANDO_APROVACAO`.

```text
AGUARDANDO_APROVACAO
  ↓ rejeitar + motivo obrigatório
REJEITADA
  ↓ editar orçamento
PATCH /orcamento/enviar-aprovacao
  ↓
AGUARDANDO_APROVACAO
```

O motivo é registrado no histórico da OS.

## Alteração da quantidade após aprovação

Não se altera `ReservaPeca.quantidade` diretamente.

Quando a OS já está `APROVADA`, a rota abaixo altera a quantidade no próprio `ItemPeca`:

```http
PATCH /v1/ordens-servico/:id/orcamento/itens-peca/:itemId/quantidade
```

Essa alteração invalida a aprovação anterior:

```text
APROVADA
  ↓ alterar quantidade de ItemPeca
BEGIN
  ↓
bloquear peças das reservas atuais
  ↓
remover reservas antigas da OS
  ↓
atualizar quantidade e totais do orçamento
  ↓
OS = AGUARDANDO_APROVACAO
  ↓
COMMIT
  ↓
reenviar orçamento atualizado ao cliente
```

Nenhuma reserva nova é criada nessa etapa. O estoque só volta a ser comprometido quando o cliente aprovar o orçamento atualizado.

```text
AGUARDANDO_APROVACAO
  ↓ nova aprovação
validar estoque novamente
  ↓
recriar reservas conforme o orçamento atual
  ↓
APROVADA
```

Isso garante que a reserva sempre reflita a versão efetivamente aprovada pelo cliente.

## Histórico da Ordem de Serviço

As decisões de aprovação/rejeição e os reenvios são registrados em `HistoricoStatus`.

Como `historicos_status.alterado_por` referencia usuários internos na estrutura atual, aprovação e rejeição realizadas pelo cliente preservam o último usuário interno do histórico como `alterado_por`. O `ClienteID` autenticado é usado para autorização e validação de propriedade da OS, não é persistido nesse campo.

## Resumo das responsabilidades

| Responsabilidade | Fonte/Componente |
|---|---|
| Quantidade solicitada | `ItemPeca` do orçamento |
| Estoque físico | `Peca.quantidadeEmEstoque` |
| Estoque mínimo | `Peca.estoqueMinimo` |
| Quantidade reservada | `ReservaPeca` / `reservas_pecas` |
| Criar reserva | `AprovarOrcamentoUseCase` |
| Alterar quantidade desejada | `AlterarQuantidadePecaOrcamentoUseCase` |
| Invalidar reserva antiga | fluxo de alteração do orçamento aprovado |
| Recriar reserva | nova aprovação do cliente |
| Proteger concorrência | transação + `FOR UPDATE` |
