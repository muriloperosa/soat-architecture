package migration

import "errors"

var (
	ErrDriverInvalido     = errors.New("driver de banco de dados inválido")
	ErrDriverNaoSuportado = errors.New("driver de banco de dados ainda não suportado para migrations")
	ErrConexaoObrigatoria = errors.New("conexão com banco de dados é obrigatória")
)
