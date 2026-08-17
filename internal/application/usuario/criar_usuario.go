package usuario

import (
	"context"
	"errors"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainusuario "github.com/muriloperosa/soat-architecture/internal/domain/usuario"
)

// CriarUsuarioUseCase cria um usuário interno com senha inicial provisória
// (definida pelo administrador, troca forçada no primeiro acesso).
type CriarUsuarioUseCase struct {
	repo domainusuario.UsuarioRepository
}

func NewCriarUsuarioUseCase(repo domainusuario.UsuarioRepository) *CriarUsuarioUseCase {
	return &CriarUsuarioUseCase{repo: repo}
}

func (uc *CriarUsuarioUseCase) Executar(ctx context.Context, input CriarUsuarioInput) (UsuarioOutput, error) {
	_, err := uc.repo.BuscarPorEmail(ctx, input.Email)
	if err == nil {
		return UsuarioOutput{}, shared.NewConflictError("já existe um usuário com esse email")
	}
	if !errors.Is(err, domainusuario.ErrUsuarioNaoEncontrado) {
		return UsuarioOutput{}, shared.NewInternalError("erro ao verificar email", err)
	}

	u, err := domainusuario.NewUsuario(input.Nome, input.Email, input.SenhaInicial, input.Papel)
	if err != nil {
		return UsuarioOutput{}, err
	}

	if err := uc.repo.Salvar(ctx, u); err != nil {
		return UsuarioOutput{}, shared.NewInternalError("erro ao salvar usuário", err)
	}

	return toOutput(u), nil
}
