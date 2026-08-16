package usuario

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// AtivarUsuarioUseCase reabilita um usuário pra login.
type AtivarUsuarioUseCase struct {
	repo domainusuario.UsuarioRepository
}

func NewAtivarUsuarioUseCase(repo domainusuario.UsuarioRepository) *AtivarUsuarioUseCase {
	return &AtivarUsuarioUseCase{repo: repo}
}

func (uc *AtivarUsuarioUseCase) Executar(ctx context.Context, id uint64) error {
	u, err := uc.repo.BuscarPorID(ctx, id)
	if err != nil {
		if errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado) {
			return err
		}
		return shared.NewInternalError("erro ao buscar usuário", err)
	}
	u.Ativar()
	if err := uc.repo.Atualizar(ctx, u); err != nil {
		return shared.NewInternalError("erro ao ativar usuário", err)
	}
	return nil
}
