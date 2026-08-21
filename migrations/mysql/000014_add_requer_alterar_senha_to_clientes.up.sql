ALTER TABLE clientes
    ADD COLUMN requer_alterar_senha BOOLEAN NOT NULL DEFAULT TRUE
    AFTER senha_hash;