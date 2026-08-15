CREATE TABLE pecas (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    codigo VARCHAR(50) NOT NULL,

    nome VARCHAR(150) NOT NULL,
    marca VARCHAR(100) NOT NULL,
    descricao VARCHAR(500) NOT NULL,

    preco DECIMAL(10,2) NOT NULL,

    quantidade_em_estoque INT UNSIGNED NOT NULL DEFAULT 0,
    quantidade_reservada INT UNSIGNED NOT NULL DEFAULT 0,
    estoque_minimo INT UNSIGNED NOT NULL DEFAULT 0,

    -- Usuário responsável pelo cadastro da peça.
    criado_por BIGINT UNSIGNED NOT NULL,

    ativo BOOLEAN NOT NULL DEFAULT TRUE,

    data_cadastro DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_pecas
        PRIMARY KEY (id),

    CONSTRAINT uk_pecas_codigo
        UNIQUE (codigo),

    CONSTRAINT fk_pecas_criado_por
        FOREIGN KEY (criado_por)
        REFERENCES usuarios(id),

    CONSTRAINT chk_pecas_preco
        CHECK (preco >= 0),

    CONSTRAINT chk_pecas_quantidade_reservada
        CHECK (quantidade_reservada <= quantidade_em_estoque)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;