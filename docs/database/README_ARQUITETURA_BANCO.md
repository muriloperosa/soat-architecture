# Arquitetura do Banco de Dados

## Visão geral

O projeto utiliza **MySQL 8** como banco de dados atual, **GORM** para
persistência e **golang-migrate** para versionamento explícito do
schema.

As migrations são SQL versionadas. O `AutoMigrate` do GORM não é
utilizado.

## Principais decisões

### IDs

As tabelas utilizam IDs internos com auto incremento:

``` sql
id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT
```

No domínio Go, os IDs podem ser representados por `uint64`.

Identificadores de negócio são separados do ID interno. Por exemplo,
`ordens_servico.numero` é gerado pela aplicação.

### Enums

Campos com conjunto fechado de valores utilizam `ENUM` no MySQL.

**Papel do usuário:**

``` sql
papel ENUM(
    'ADMINISTRADOR',
    'ATENDENTE',
    'MECANICO'
) NOT NULL
```

**Tipo de pessoa:**

``` sql
tipo ENUM('PF', 'PJ') NOT NULL
```

**Status da OS:**

``` sql
status ENUM(
    'RECEBIDA',
    'EM_DIAGNOSTICO',
    'AGUARDANDO_APROVACAO',
    'APROVADA',
    'REJEITADA',
    'EM_EXECUCAO',
    'FINALIZADA',
    'ENTREGUE'
) NOT NULL
```

Os mesmos valores de status são usados em
`historicos_status.status_novo`.

O domínio Go também deve possuir seus próprios tipos/enums, mantendo as
regras protegidas na aplicação e no banco.

### Senha

Tanto usuários quanto clientes armazenam apenas o **hash** da senha:

``` sql
senha_hash VARCHAR(255) NOT NULL
```

Senha em texto puro nunca deve ser persistida.

### Cliente e veículo

Cliente e veículo não possuem relacionamento direto.

O vínculo histórico é determinado pela Ordem de Serviço:

``` text
Cliente ──> Ordem de Serviço <── Veículo
```

Isso permite que um mesmo veículo apareça historicamente em ordens de
serviço de clientes diferentes, como em um cenário de venda do veículo.

### Quilometragem

`veiculos.quilometragem_atual` representa a última quilometragem
conhecida.

`ordens_servico.quilometragem_entrada` preserva a quilometragem
existente no momento da abertura da OS.

Assim, atualizar o veículo não altera o histórico das OS anteriores.

### Orçamento

Uma Ordem de Serviço possui no máximo um orçamento.

Se o orçamento for rejeitado, ele pode ser editado e enviado novamente
para aprovação. A aprovação/rejeição é representada pelo status da OS.

Os itens preservam valores históricos para que alterações posteriores
nos cadastros de serviços e peças não modifiquem o orçamento existente.

### Auditoria

`criado_por` é utilizado somente onde existe uma ação relevante de
criação:

-   clientes;
-   veículos;
-   serviços;
-   peças;
-   ordens de serviço;
-   orçamentos.

`historicos_status` utiliza `alterado_por`.

Tabelas internas, como itens do orçamento e reservas, não recebem
`criado_por` individualmente.

## Estoque e reservas

### Fonte de verdade

A tabela `pecas` mantém o estoque físico:

``` sql
quantidade_em_estoque INT UNSIGNED NOT NULL DEFAULT 0,
estoque_minimo INT UNSIGNED NOT NULL DEFAULT 0
```

Não existe `quantidade_reservada` em `pecas`.

As reservas ficam exclusivamente em `reservas_pecas`, evitando
duplicação do mesmo estado.

A disponibilidade é:

``` text
quantidade_em_estoque
- total reservado
= quantidade disponível
```

### Estrutura da reserva

Uma reserva identifica a OS, a peça e a quantidade:

``` sql
CREATE TABLE reservas_pecas (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    ordem_servico_id BIGINT UNSIGNED NOT NULL,
    peca_id BIGINT UNSIGNED NOT NULL,
    quantidade INT UNSIGNED NOT NULL,

    criada_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    atualizada_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_reservas_pecas PRIMARY KEY (id),

    CONSTRAINT fk_reservas_pecas_ordem_servico
        FOREIGN KEY (ordem_servico_id)
        REFERENCES ordens_servico(id),

    CONSTRAINT fk_reservas_pecas_peca
        FOREIGN KEY (peca_id)
        REFERENCES pecas(id),

    CONSTRAINT uk_reservas_pecas_ordem_peca
        UNIQUE (ordem_servico_id, peca_id),

    CONSTRAINT chk_reservas_pecas_quantidade
        CHECK (quantidade > 0)
);
```

A combinação `ordem_servico_id + peca_id` é única. Se a quantidade
reservada mudar, a reserva existente deve ser atualizada.

## Exemplo de consulta da disponibilidade

``` sql
SELECT
    P.id,
    P.quantidade_em_estoque,
    P.estoque_minimo,
    COALESCE(SUM(R.quantidade), 0) AS quantidade_reservada,
    P.quantidade_em_estoque - COALESCE(SUM(R.quantidade), 0)
        AS quantidade_disponivel
FROM pecas P
LEFT JOIN reservas_pecas R
       ON R.peca_id = P.id
WHERE P.id = ?
GROUP BY
    P.id,
    P.quantidade_em_estoque,
    P.estoque_minimo;
```

## Reserva transacional

A validação de disponibilidade e a alteração da reserva devem ocorrer na
**mesma transação**.

Exemplo conceitual MySQL:

``` sql
START TRANSACTION;

-- Bloqueia a peça durante a operação.
SELECT
    id,
    quantidade_em_estoque,
    estoque_minimo
FROM pecas
WHERE id = ?
FOR UPDATE;

-- Obtém o total atualmente reservado.
SELECT
    COALESCE(SUM(quantidade), 0) AS quantidade_reservada
FROM reservas_pecas
WHERE peca_id = ?;
```

A aplicação calcula:

``` text
saldo_apos_reserva =
    quantidade_em_estoque
    - quantidade_reservada_atual
    - nova_quantidade
```

A reserva só pode continuar quando:

``` text
saldo_apos_reserva >= estoque_minimo
```

Caso contrário:

``` sql
ROLLBACK;
```

Caso seja válida, a reserva pode ser inserida/atualizada:

``` sql
INSERT INTO reservas_pecas (
    ordem_servico_id,
    peca_id,
    quantidade
)
VALUES (?, ?, ?)
ON DUPLICATE KEY UPDATE
    quantidade = VALUES(quantidade),
    atualizada_em = CURRENT_TIMESTAMP;

COMMIT;
```

> A regra final deve ser implementada pela aplicação dentro de uma
> transação. O exemplo SQL documenta o comportamento esperado.

## Exemplo de remoção de reserva

Para remover uma reserva específica:

``` sql
DELETE FROM reservas_pecas
WHERE ordem_servico_id = ?
  AND peca_id = ?;
```

A operação também deve fazer parte da transação do caso de uso
correspondente quando houver outras alterações relacionadas.

## Exemplo de reservas de uma OS

``` sql
SELECT
    R.id,
    R.peca_id,
    P.codigo,
    P.nome,
    R.quantidade
FROM reservas_pecas R
INNER JOIN pecas P
        ON P.id = R.peca_id
WHERE R.ordem_servico_id = ?;
```

## Migrations

### Estratégia

O projeto utiliza `golang-migrate`.

Não é utilizado:

``` go
db.AutoMigrate(...)
```

Cada mudança estrutural deve possuir migration `up` e `down`.

Exemplo:

``` text
migrations/
└── mysql/
    ├── 000001_create_usuarios.up.sql
    ├── 000001_create_usuarios.down.sql
    ├── 000002_create_clientes.up.sql
    ├── 000002_create_clientes.down.sql
    └── ...
```

As migrations são executadas pela numeração. Portanto, dependências
devem respeitar a ordem dos arquivos.

Por exemplo, `usuarios` precisa existir antes das tabelas que possuem FK
`criado_por`.

### Up

Exemplo:

``` sql
CREATE TABLE exemplo (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    nome VARCHAR(150) NOT NULL,

    CONSTRAINT pk_exemplo PRIMARY KEY (id)
);
```

### Down

``` sql
DROP TABLE IF EXISTS exemplo;
```

O `down` deve desfazer o que foi realizado pelo respectivo `up`.

## Migrations específicas por SGBD

As migrations atuais utilizam recursos específicos do MySQL, como:

-   `AUTO_INCREMENT`;
-   `UNSIGNED`;
-   `ENUM`;
-   `ENGINE=InnoDB`.

Caso outro banco seja suportado, ele poderá possuir migrations próprias:

``` text
migrations/
├── mysql/
│   └── ...
└── postgres/
    └── ...
```

O `DriverType` já reconhece os tipos planejados, mas atualmente apenas
MySQL está efetivamente implementado para migrations.

A abstração da **conexão da aplicação** por `DriverType` ainda é uma
melhoria futura.

## Comandos disponíveis

### Subir somente o MySQL

``` bash
make db-up
```

### Parar o MySQL

``` bash
make db-down
```

### Subir banco e executar migrations

``` bash
make db-setup
```

Equivale conceitualmente a:

``` text
subir MySQL
    ↓
aguardar healthcheck
    ↓
executar migrations pendentes
```

### Aplicar migrations pendentes

``` bash
make migrate-up
```

Executa:

``` bash
go run ./migrations up
```

### Desfazer a última migration

``` bash
make migrate-down
```

Executa:

``` bash
go run ./migrations down
```

### Consultar versão

``` bash
make migrate-version
```

Exemplo esperado após todas as migrations:

``` text
versão atual: 11 | dirty: false
```

### Corrigir versão dirty

Quando uma migration falha no meio da execução, o `golang-migrate` pode
marcar o banco como `dirty`.

Use `force` somente após verificar/corrigir manualmente o estado do
schema:

``` bash
go run ./migrations force <VERSAO>
```

Depois:

``` bash
make migrate-up
```

`force` não executa nem desfaz SQL; ele apenas altera a versão
registrada pelo `golang-migrate`.

### Resetar banco local

Durante o desenvolvimento, para apagar o volume e executar tudo
novamente:

``` bash
make db-reset
```

Fluxo esperado:

``` text
remover containers/volume
        ↓
criar MySQL vazio
        ↓
aguardar banco saudável
        ↓
executar migration 000001
        ↓
...
        ↓
executar última migration
```

> `db-reset` remove os dados locais. Não deve ser utilizado em produção.

## Fluxo recomendado ao criar uma migration

1.  Criar os arquivos `up` e `down`.
2.  Garantir que a numeração seja posterior à última migration.
3.  Implementar o SQL de criação/alteração no `up`.
4.  Implementar a reversão correspondente no `down`.
5.  Executar:

``` bash
make migrate-up
```

6.  Conferir:

``` bash
make migrate-version
```

7.  Testar rollback:

``` bash
make migrate-down
```

8.  Aplicar novamente:

``` bash
make migrate-up
```

Durante a construção inicial do schema, também é útil validar tudo
partindo do zero:

``` bash
make db-reset
```

## Regras importantes

-   Não utilizar `AutoMigrate`.
-   Toda alteração de schema deve possuir migration.
-   Não armazenar senha em texto puro.
-   Não criar relacionamento direto entre cliente e veículo.
-   Preservar dados históricos nos itens do orçamento.
-   Não duplicar quantidade reservada em `pecas`.
-   Consultar reservas através de `reservas_pecas`.
-   Reserva/consumo de estoque deve ser transacional.
-   A operação não pode deixar o estoque abaixo de `estoque_minimo`.
-   Usar `FOR UPDATE` no fluxo de reserva para proteger contra
    concorrência.
-   `force` deve ser utilizado somente para recuperação controlada de
    migrations `dirty`.
-   `db-reset` é exclusivo para ambientes descartáveis/de
    desenvolvimento.

## Melhorias futuras

-   Abstrair a criação da conexão da aplicação através de `DriverType`.
-   Implementar suporte real a outros SGBDs somente quando necessário.
-   Criar migrations específicas para cada SGBD suportado.
-   Adicionar testes de integração com MySQL real.
-   Testar concorrência nas operações de reserva/consumo de estoque.
