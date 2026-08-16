package migration

import (
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

var (
	ErrDriverInvalido     = shared.NewInternalErrorCustom("driver de banco de dados inválido")
	ErrConexaoObrigatoria = shared.NewInternalErrorCustom("conexão com banco de dados é obrigatória")
	ErrDriverNaoSuportado = shared.NewInternalErrorCustom("driver de banco de dados ainda não suportado para migrations")
)
