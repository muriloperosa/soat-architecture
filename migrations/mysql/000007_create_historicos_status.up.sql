CREATE TABLE historicos_status (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    ordem_servico_id BIGINT UNSIGNED NOT NULL,

    status ENUM(
        'RECEBIDA',
        'EM_DIAGNOSTICO',
        'AGUARDANDO_APROVACAO',
        'APROVADA',
        'REJEITADA',
        'EM_EXECUCAO',
        'FINALIZADA',
        'ENTREGUE'
    ) NOT NULL,

    alterado_por BIGINT UNSIGNED NOT NULL,

    motivo VARCHAR(500) NULL,

    alterado_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_historicos_status
        PRIMARY KEY (id),

    CONSTRAINT fk_historicos_status_ordem_servico
        FOREIGN KEY (ordem_servico_id)
        REFERENCES ordens_servico(id),

    CONSTRAINT fk_historicos_status_usuario
        FOREIGN KEY (alterado_por)
        REFERENCES usuarios(id),

    INDEX idx_historicos_status_ordem_servico_data (
        ordem_servico_id,
        alterado_em
    )
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;