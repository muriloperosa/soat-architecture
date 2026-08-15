CREATE TABLE usuarios (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    papel VARCHAR(20) NOT NULL,
    nome VARCHAR(150) NOT NULL,
    email VARCHAR(255) NOT NULL,
    senha_hash VARCHAR(255) NOT NULL,

    ativo BOOLEAN NOT NULL DEFAULT TRUE,

    data_cadastro DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_usuarios
        PRIMARY KEY (id),

    CONSTRAINT uk_usuarios_email
        UNIQUE (email),

    CONSTRAINT chk_usuarios_papel
        CHECK (
            papel IN (
                'ADMINISTRADOR',
                'ATENDENTE',
                'MECANICO'
            )
        )
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;