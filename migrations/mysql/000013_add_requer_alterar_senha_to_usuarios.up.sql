-- true ate o usuario trocar a senha definida pelo administrador no
-- primeiro acesso (troca forçada).
ALTER TABLE usuarios
    ADD COLUMN requer_alterar_senha BOOLEAN NOT NULL DEFAULT TRUE AFTER senha_hash;
