package cliente

import (
	"context"
	"errors"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

type CadastrarInput struct {
	Documento string
	Tipo      domain.TipoPessoa
	Nome      string
	Email     string
	Telefone  string
	Senha     string
}

type Cadastrar struct {
	repository domain.Repository
}

func NewCadastrar(repository domain.Repository) *Cadastrar {
	return &Cadastrar{repository: repository}
}

func (uc *Cadastrar) Executar(ctx context.Context, input CadastrarInput) (*domain.Cliente, error) {
	existente, err := uc.repository.BuscarPorDocumento(ctx, input.Documento)
	if err != nil && !errors.Is(err, domain.ErrClienteNaoEncontrado) {
		return nil, err
	}

	if existente != nil {
		return nil, domain.ErrClienteJaCadastrado
	}

	cliente, err := domain.NewCliente(
		input.Documento,
		input.Tipo,
		input.Nome,
		input.Email,
		input.Telefone,
		input.Senha,
	)
	if err != nil {
		return nil, err
	}

	if err := uc.repository.Salvar(ctx, &cliente); err != nil {
		return nil, err
	}

	return &cliente, nil
}
