CREATE TABLE reservas_pecas (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    -- Ordem de Serviço que originou a reserva.
    ordem_servico_id BIGINT UNSIGNED NOT NULL,

    -- Peça reservada para a OS.
    peca_id BIGINT UNSIGNED NOT NULL,

    -- Quantidade atualmente reservada desta peça para esta OS.
    quantidade INT UNSIGNED NOT NULL,

    criada_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    atualizada_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_reservas_pecas
        PRIMARY KEY (id),

    CONSTRAINT fk_reservas_pecas_ordem_servico
        FOREIGN KEY (ordem_servico_id)
        REFERENCES ordens_servico(id),

    CONSTRAINT fk_reservas_pecas_peca
        FOREIGN KEY (peca_id)
        REFERENCES pecas(id),

    -- Para o MVP, cada peça aparece no máximo uma vez
    -- como reserva dentro da mesma OS.
    -- Se a quantidade mudar, atualizamos o mesmo registro.
    CONSTRAINT uk_reservas_pecas_ordem_peca
        UNIQUE (ordem_servico_id, peca_id),

    CONSTRAINT chk_reservas_pecas_quantidade
        CHECK (quantidade > 0),

    -- Facilita consultas de todas as reservas de uma OS.
    INDEX idx_reservas_pecas_ordem_servico (
        ordem_servico_id
    ),

    -- Facilita consultas de reservas existentes para uma peça.
    INDEX idx_reservas_pecas_peca (
        peca_id
    )
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;