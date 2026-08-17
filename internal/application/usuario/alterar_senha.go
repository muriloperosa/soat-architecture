package usuario

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// AlterarSenhaUseCase troca a senha do próprio usuário logado,
// usado tanto pra troca voluntária quanto pra destravar o primeiro acesso.
type AlterarSenhaUseCase struct {
	repo domainusuario.UsuarioRepository
}

func NewAlterarSenhaUseCase(repo domainusuario.UsuarioRepository) *AlterarSenhaUseCase {
	return &AlterarSenhaUseCase{repo: repo}
}

func (uc *AlterarSenhaUseCase) Executar(ctx context.Context, input AlterarSenhaInput) error {
	u, err := uc.repo.BuscarPorID(ctx, input.UsuarioID)
	if err != nil {
		if errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado) {
			return err
		}
		return shared.NewInternalError("erro ao buscar usuário", err)
	}

	if err := u.AlterarSenha(input.SenhaNova); err != nil {
		return err
	}

	if err := uc.repo.Atualizar(ctx, u); err != nil {
		return shared.NewInternalError("erro ao atualizar senha", err)
	}
	return nil
}
