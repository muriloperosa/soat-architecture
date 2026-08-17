package usuario

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// AtualizarUsuarioUseCase troca nome, email e papel de um usuário existente,
// e opcionalmente redefine a senha (admin).
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

	if input.Email != u.Email().String() {
		outro, err := uc.repo.BuscarPorEmail(ctx, input.Email)
		if err == nil && outro.ID() != u.ID() {
			return UsuarioOutput{}, shared.NewConflictError("já existe um usuário com esse email")
		}
		if err != nil && !errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado) {
			return UsuarioOutput{}, shared.NewInternalError("erro ao verificar email", err)
		}
	}

	if err := u.Atualizar(input.Nome, input.Email, input.Papel); err != nil {
		return UsuarioOutput{}, err
	}

	if input.SenhaNova != "" {
		if err := u.RedefinirSenha(input.SenhaNova); err != nil {
			return UsuarioOutput{}, err
		}
	}

	if err := uc.repo.Atualizar(ctx, u); err != nil {
		return UsuarioOutput{}, shared.NewInternalError("erro ao atualizar usuário", err)
	}

	return toOutput(u), nil
}
