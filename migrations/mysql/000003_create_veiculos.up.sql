CREATE TABLE veiculos (
    id BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,

    placa VARCHAR(7) NOT NULL,
    marca VARCHAR(100) NOT NULL,
    modelo VARCHAR(100) NOT NULL,
    quilometragem_atual INT UNSIGNED NOT NULL,
    ano SMALLINT UNSIGNED NOT NULL,
    cor VARCHAR(50) NOT NULL,

    -- Usuário responsável pelo cadastro do veículo.
    criado_por BIGINT UNSIGNED NOT NULL,

    ativo BOOLEAN NOT NULL DEFAULT TRUE,

    data_cadastro DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    data_atualizacao DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
        ON UPDATE CURRENT_TIMESTAMP,

    CONSTRAINT pk_veiculos
        PRIMARY KEY (id),

    CONSTRAINT uk_veiculos_placa
        UNIQUE (placa),

    CONSTRAINT fk_veiculos_criado_por
        FOREIGN KEY (criado_por)
        REFERENCES usuarios(id),

    CONSTRAINT chk_veiculos_quilometragem
        CHECK (quilometragem_atual >= 0),

    CONSTRAINT chk_veiculos_ano
        CHECK (ano >= 1900)
) ENGINE=InnoDB
  DEFAULT CHARSET=utf8mb4
  COLLATE=utf8mb4_unicode_ci;