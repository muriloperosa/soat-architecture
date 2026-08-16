package mysql

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

var (
	ErrConfigNotFound = shared.NewInternalErrorCustom("configuração é obrigatória")
)
