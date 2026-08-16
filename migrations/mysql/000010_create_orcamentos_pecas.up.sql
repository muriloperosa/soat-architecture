CREATE TABLE orcamentos_pecas (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    -- Orçamento ao qual este item pertence.
    orcamento_id BIGINT UNSIGNED NOT NULL,

    -- Peça de origem utilizada como referência.
    peca_id BIGINT UNSIGNED NOT NULL,

    -- Descrição preservada no momento do orçamento.
    -- Isso evita que alterações futuras no cadastro da peça
    -- modifiquem o histórico do orçamento.
    descricao VARCHAR(500) NOT NULL,

    -- Quantidade de unidades da peça incluídas no orçamento.
    quantidade INT UNSIGNED NOT NULL,

    -- Valor unitário aplicado no orçamento.
    -- É preservado historicamente e não acompanha mudanças
    -- futuras em pecas.preco.
    valor DECIMAL(10,2) NOT NULL,

    CONSTRAINT pk_orcamentos_pecas
        PRIMARY KEY (id),

    CONSTRAINT fk_orcamentos_pecas_orcamento
        FOREIGN KEY (orcamento_id)
        REFERENCES orcamentos(id),

    CONSTRAINT fk_orcamentos_pecas_peca
        FOREIGN KEY (peca_id)
        REFERENCES pecas(id),

    CONSTRAINT chk_orcamentos_pecas_quantidade
        CHECK (quantidade > 0),

    CONSTRAINT chk_orcamentos_pecas_valor
        CHECK (valor >= 0)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;