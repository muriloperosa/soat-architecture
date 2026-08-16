CREATE TABLE servicos (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    nome VARCHAR(150) NOT NULL,
    descricao VARCHAR(500) NOT NULL,

    preco_base DECIMAL(10,2) NOT NULL,
    tempo_estimado_minutos INT UNSIGNED NOT NULL,

    -- Usuário responsável pelo cadastro do serviço.
    criado_por BIGINT UNSIGNED NOT NULL,

    ativo BOOLEAN NOT NULL DEFAULT TRUE,

    data_cadastro DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_servicos
        PRIMARY KEY (id),

    CONSTRAINT fk_servicos_criado_por
        FOREIGN KEY (criado_por)
        REFERENCES usuarios(id),

    CONSTRAINT chk_servicos_preco_base
        CHECK (preco_base >= 0),

    CONSTRAINT chk_servicos_tempo_estimado
        CHECK (tempo_estimado_minutos > 0)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;