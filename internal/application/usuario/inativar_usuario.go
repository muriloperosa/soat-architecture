package usuario

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// InativarUsuarioUseCase bloqueia um usuário de logar (LoginUseCase rejeita
// via Credencial.Ativo).
type InativarUsuarioUseCase struct {
	repo domainusuario.UsuarioRepository
}

func NewInativarUsuarioUseCase(repo domainusuario.UsuarioRepository) *InativarUsuarioUseCase {
	return &InativarUsuarioUseCase{repo: repo}
}

func (uc *InativarUsuarioUseCase) Executar(ctx context.Context, id uint64) error {
	u, err := uc.repo.BuscarPorID(ctx, id)
	if err != nil {
		if errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado) {
			return err
		}
		return shared.NewInternalError("erro ao buscar usuário", err)
	}
	u.Inativar()
	if err := uc.repo.Atualizar(ctx, u); err != nil {
		return shared.NewInternalError("erro ao inativar usuário", err)
	}
	return nil
}
