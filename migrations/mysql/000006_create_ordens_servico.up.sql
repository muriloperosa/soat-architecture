CREATE TABLE ordens_servico (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    numero VARCHAR(50) NOT NULL,

    cliente_id BIGINT UNSIGNED NOT NULL,
    veiculo_id BIGINT UNSIGNED NOT NULL,

    quilometragem_entrada INT UNSIGNED NOT NULL,

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

    diagnostico TEXT NULL,
    observacoes TEXT NULL,

    criado_por BIGINT UNSIGNED NOT NULL,

    data_cadastro DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_ordens_servico
        PRIMARY KEY (id),

    CONSTRAINT uk_ordens_servico_numero
        UNIQUE (numero),

    CONSTRAINT fk_ordens_servico_cliente
        FOREIGN KEY (cliente_id)
        REFERENCES clientes(id),

    CONSTRAINT fk_ordens_servico_veiculo
        FOREIGN KEY (veiculo_id)
        REFERENCES veiculos(id),

    CONSTRAINT fk_ordens_servico_criado_por
        FOREIGN KEY (criado_por)
        REFERENCES usuarios(id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;