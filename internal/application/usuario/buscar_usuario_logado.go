package usuario

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// BuscarUsuarioLogadoUseCase consulta os dados básicos do usuário
// autenticado (GET /v1/usuarios/me).
type BuscarUsuarioLogadoUseCase struct {
	repo domainusuario.UsuarioRepository
}

func NewBuscarUsuarioLogadoUseCase(repo domainusuario.UsuarioRepository) *BuscarUsuarioLogadoUseCase {
	return &BuscarUsuarioLogadoUseCase{repo: repo}
}

func (uc *BuscarUsuarioLogadoUseCase) Executar(ctx context.Context, id uint64) (UsuarioOutput, error) {
	u, err := uc.repo.BuscarPorID(ctx, id)
	if err != nil {
		if errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado) {
			return UsuarioOutput{}, err
		}
		return UsuarioOutput{}, shared.NewInternalError("erro ao buscar usuário", err)
	}
	return toOutput(u), nil
}
