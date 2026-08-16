package usuario

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// AtualizarUsuarioUseCase troca nome e papel de um usuário existente.
type AtualizarUsuarioUseCase struct {
	repo domainusuario.UsuarioRepository
}

func NewAtualizarUsuarioUseCase(repo domainusuario.UsuarioRepository) *AtualizarUsuarioUseCase {
	return &AtualizarUsuarioUseCase{repo: repo}
}

func (uc *AtualizarUsuarioUseCase) Executar(ctx context.Context, input AtualizarUsuarioInput) (UsuarioOutput, error) {
	u, err := uc.repo.BuscarPorID(ctx, input.ID)
	if err != nil {
		if errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado) {
			return UsuarioOutput{}, err
		}
		return UsuarioOutput{}, shared.NewInternalError("erro ao buscar usuário", err)
	}

	if err := u.Atualizar(input.Nome, input.Papel); err != nil {
		return UsuarioOutput{}, err
	}

	if err := uc.repo.Atualizar(ctx, u); err != nil {
		return UsuarioOutput{}, shared.NewInternalError("erro ao atualizar usuário", err)
	}

	return toOutput(u), nil
}
