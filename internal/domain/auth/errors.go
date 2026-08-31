package auth

import "github.com/muriloperosa/soat-architecture/internal/domain/shared"

// ErrRefreshTokenNaoEncontrado indica que nenhum refresh token corresponde ao hash buscado.
var ErrRefreshTokenNaoEncontrado = shared.NewNotFoundError("refresh token não encontrado")
