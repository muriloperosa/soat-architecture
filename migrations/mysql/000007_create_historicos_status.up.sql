CREATE TABLE historicos_status (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    -- Ordem de Serviço que teve o status alterado.
    ordem_servico_id BIGINT UNSIGNED NOT NULL,

    -- Novo status assumido pela OS.
    -- O status anterior pode ser obtido pelo registro
    -- imediatamente anterior no histórico da mesma OS.
    status_novo VARCHAR(30) NOT NULL,

    -- Usuário responsável pela alteração.
    alterado_por BIGINT UNSIGNED NOT NULL,

    -- Motivo opcional da alteração.
    -- É especialmente útil em situações como rejeição da OS.
    motivo VARCHAR(500) NULL,

    -- Momento em que a alteração ocorreu.
    alterado_em DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT pk_historicos_status
        PRIMARY KEY (id),

    CONSTRAINT fk_historicos_status_ordem_servico
        FOREIGN KEY (ordem_servico_id)
        REFERENCES ordens_servico(id),

    CONSTRAINT fk_historicos_status_usuario
        FOREIGN KEY (alterado_por)
        REFERENCES usuarios(id),

    CONSTRAINT chk_historicos_status_status
        CHECK (
            status_novo IN (
                'RECEBIDA',
                'EM_DIAGNOSTICO',
                'AGUARDANDO_APROVACAO',
                'APROVADA',
                'REJEITADA',
                'EM_EXECUCAO',
                'FINALIZADA',
                'ENTREGUE'
            )
        ),

    -- Facilita consultas do histórico de uma OS em ordem cronológica.
    INDEX idx_historicos_status_ordem_servico_data (
        ordem_servico_id,
        alterado_em
    )
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;