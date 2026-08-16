CREATE TABLE orcamentos (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    ordem_servico_id BIGINT UNSIGNED NOT NULL,

    valor_item_servicos DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    valor_item_pecas DECIMAL(10,2) NOT NULL DEFAULT 0.00,
    valor_total DECIMAL(10,2) NOT NULL DEFAULT 0.00,

    observacoes VARCHAR(500) NULL,

    -- Usuário responsável pela criação inicial do orçamento.
    -- O mesmo orçamento pode ser editado posteriormente.
    criado_por BIGINT UNSIGNED NOT NULL,

    criado_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    atualizado_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_orcamentos
        PRIMARY KEY (id),

    -- Uma OS possui no máximo um orçamento.
    CONSTRAINT uk_orcamentos_ordem_servico
        UNIQUE (ordem_servico_id),

    CONSTRAINT fk_orcamentos_ordem_servico
        FOREIGN KEY (ordem_servico_id)
        REFERENCES ordens_servico(id),

    CONSTRAINT fk_orcamentos_criado_por
        FOREIGN KEY (criado_por)
        REFERENCES usuarios(id),

    CONSTRAINT chk_orcamentos_valor_servicos
        CHECK (valor_item_servicos >= 0),

    CONSTRAINT chk_orcamentos_valor_pecas
        CHECK (valor_item_pecas >= 0),

    CONSTRAINT chk_orcamentos_valor_total
        CHECK (valor_total >= 0)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;