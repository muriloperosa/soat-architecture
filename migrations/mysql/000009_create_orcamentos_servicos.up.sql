CREATE TABLE orcamentos_servicos (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    -- Orçamento ao qual este item pertence.
    orcamento_id BIGINT UNSIGNED NOT NULL,

    -- Serviço de origem utilizado como referência.
    servico_id BIGINT UNSIGNED NOT NULL,

    -- Quantidade de unidades do serviço incluídas no orçamento.
    -- Deve ser maior que zero.
    quantidade INT UNSIGNED NOT NULL,

    -- Valor unitário aplicado neste orçamento.
    --
    -- Esse valor é preservado no item para que alterações futuras
    -- em servicos.preco_base não modifiquem orçamentos antigos.
    valor DECIMAL(10,2) NOT NULL,

    -- Duração estimada aplicada ao item no momento do orçamento.
    -- Também é preservada historicamente, independentemente de
    -- alterações posteriores no cadastro do serviço.
    tempo_estimado_minutos INT UNSIGNED NOT NULL,

    CONSTRAINT pk_orcamentos_servicos
        PRIMARY KEY (id),

    CONSTRAINT fk_orcamentos_servicos_orcamento
        FOREIGN KEY (orcamento_id)
        REFERENCES orcamentos(id),

    CONSTRAINT fk_orcamentos_servicos_servico
        FOREIGN KEY (servico_id)
        REFERENCES servicos(id),

    CONSTRAINT chk_orcamentos_servicos_quantidade
        CHECK (quantidade > 0),

    CONSTRAINT chk_orcamentos_servicos_valor
        CHECK (valor >= 0),

    CONSTRAINT chk_orcamentos_servicos_tempo_estimado
        CHECK (tempo_estimado_minutos > 0)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;