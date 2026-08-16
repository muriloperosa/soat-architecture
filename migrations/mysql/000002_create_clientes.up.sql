CREATE TABLE clientes (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    documento VARCHAR(14) NOT NULL,

    tipo ENUM(
        'PF',
        'PJ'
    ) NOT NULL,

    nome VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL,
    telefone VARCHAR(11) NOT NULL,
    senha_hash VARCHAR(255) NOT NULL,

    criado_por BIGINT UNSIGNED NOT NULL,

    ativo BOOLEAN NOT NULL DEFAULT TRUE,

    data_cadastro DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_clientes PRIMARY KEY (id),

    CONSTRAINT uk_clientes_documento UNIQUE (documento),

    CONSTRAINT uk_clientes_email UNIQUE (email),

    CONSTRAINT fk_clientes_criado_por
        FOREIGN KEY (criado_por)
        REFERENCES usuarios(id)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;