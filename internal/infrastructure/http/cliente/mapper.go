package cliente

import (
	app "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
)

func toCriarInput(criadoPor uint64, req CriarClienteRequest) app.CriarClienteInput {
	return app.CriarClienteInput{
		Documento: req.Documento,
		Tipo:      domain.TipoPessoa(req.TipoPessoa),
		Nome:      req.Nome,
		Email:     req.Email,
		Telefone:  req.Telefone,
		Senha:     req.Senha,
		CriadoPor: criadoPor,
	}
}

func toAtualizarInput(id uint64, req AtualizarClienteRequest) app.AtualizarClienteInput {
	return app.AtualizarClienteInput{
		ID:       id,
		Nome:     req.Nome,
		Email:    req.Email,
		Telefone: req.Telefone,
	}
}

func (r AlterarSenhaRequest) toInput(id uint64) app.AlterarSenhaInput {
	return app.AlterarSenhaInput{
		ClienteID: id,
		SenhaNova: r.SenhaNova,
	}
}

func toResponse(output app.ClienteOutput) ClienteResponse {
	return ClienteResponse{
		ID:                 output.ID,
		Documento:          output.Documento,
		TipoPessoa:         string(output.Tipo),
		Nome:               output.Nome,
		Email:              output.Email,
		Telefone:           output.Telefone,
		Ativo:              output.Ativo,
		RequerAlterarSenha: output.RequerAlterarSenha,
		CriadoPor:          output.CriadoPor,
	}
}
