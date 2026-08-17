CREATE TABLE refresh_tokens (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    -- Dono do token: aponta pra usuarios.id (tipo = 'interno') ou
    -- clientes.id (tipo = 'cliente') -- referência polimórfica, sem FK.
    usuario_id BIGINT UNSIGNED NOT NULL,

    tipo ENUM(
        'interno',
        'cliente'
    ) NOT NULL,

    papel VARCHAR(20) NOT NULL,

    token_hash VARCHAR(64) NOT NULL,

    -- Jti do access token emitido junto (mesma chamada de gerarTokens) --
    -- permite revogar os dois em par (logout ou rotação) sem esperar o
    -- access token expirar sozinho.
    access_token_jti VARCHAR(43) NOT NULL,

    expira_em DATETIME NOT NULL,
    revogado_em DATETIME NULL,

    CONSTRAINT pk_refresh_tokens
        PRIMARY KEY (id),

    CONSTRAINT uk_refresh_tokens_token_hash
        UNIQUE (token_hash),

    -- Facilita consultas de tokens ativos de um usuário.
    INDEX idx_refresh_tokens_usuario_id (
        usuario_id
    ),

    -- Usado por AuthenticationMiddleware pra checar se o access token
    -- pareado foi revogado.
    INDEX idx_refresh_tokens_access_token_jti (
        access_token_jti
    )
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;