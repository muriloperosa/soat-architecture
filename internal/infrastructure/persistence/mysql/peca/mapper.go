package peca

import domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"

func toModel(peca *domain.Peca) *PecaModel {
	return &PecaModel{
		ID:                  peca.ID(),
		Codigo:              peca.Codigo(),
		Nome:                peca.Nome(),
		Marca:               peca.Marca(),
		Descricao:           peca.Descricao(),
		Preco:               peca.Preco(),
		QuantidadeEmEstoque: peca.QuantidadeEmEstoque(),
		EstoqueMinimo:       peca.EstoqueMinimo(),
		CriadoPor:           peca.CriadoPor(),
		Ativo:               peca.Ativo(),
		DataCadastro:        peca.DataCadastro(),
		DataAtualizacao:     peca.DataAtualizacao(),
	}
}

func toDomain(model PecaModel) *domain.Peca {
	return domain.RestaurarPeca(
		model.ID,
		model.Codigo,
		model.Nome,
		model.Marca,
		model.Descricao,
		model.Preco,
		model.QuantidadeEmEstoque,
		model.EstoqueMinimo,
		model.CriadoPor,
		model.Ativo,
		model.DataCadastro,
		model.DataAtualizacao,
	)
}
