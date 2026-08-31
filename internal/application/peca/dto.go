package peca

import (
	appquery "github.com/muriloperosa/soat-architecture/internal/application/query"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
)

// CadastrarPecaInput é o DTO de entrada do CadastrarPecaUseCase.
type CadastrarPecaInput struct {
	Nome                string
	Marca               string
	Descricao           string
	Preco               float64
	QuantidadeEmEstoque int
	EstoqueMinimo       int
	CriadoPor           uint64
}

// AtualizarPecaInput é o DTO de entrada do AtualizarPecaUseCase.
type AtualizarPecaInput struct {
	ID            uint64
	Nome          string
	Marca         string
	Descricao     string
	Preco         float64
	EstoqueMinimo int
}

// ConsumirEstoqueInput é o DTO de entrada do ConsumirEstoqueUseCase.
type ConsumirEstoqueInput struct {
	PecaID     uint64
	Quantidade int
}

// ReporEstoqueInput é o DTO de entrada do ReporEstoqueUseCase.
type ReporEstoqueInput struct {
	PecaID     uint64
	Quantidade int
}

// PecaOutput é o DTO de saída comum aos use cases de gestão/consulta de peça.
type PecaOutput struct {
	ID                  uint64
	Codigo              string
	Nome                string
	Marca               string
	Descricao           string
	Preco               float64
	QuantidadeEmEstoque int
	EstoqueMinimo       int
	CriadoPor           uint64
	Ativo               bool
}

// DisponibilidadePecaOutput é o DTO de saída do ConsultarDisponibilidadeUseCase.
// QuantidadeDisponivel = QuantidadeEmEstoque - QuantidadeReservada.
type DisponibilidadePecaOutput struct {
	PecaID               uint64
	QuantidadeEmEstoque  int
	QuantidadeReservada  int
	QuantidadeDisponivel int
}

func toOutput(p *domain.Peca) PecaOutput {
	return PecaOutput{
		ID:                  p.ID(),
		Codigo:              p.Codigo(),
		Nome:                p.Nome(),
		Marca:               p.Marca(),
		Descricao:           p.Descricao(),
		Preco:               p.Preco(),
		QuantidadeEmEstoque: p.QuantidadeEmEstoque(),
		EstoqueMinimo:       p.EstoqueMinimo(),
		CriadoPor:           p.CriadoPor(),
		Ativo:               p.Ativo(),
	}
}

// FiltroPecasInput representa um filtro aceito pelo caso de uso de listagem.
type FiltroPecasInput struct {
	Field    string
	Operator string
	Value    string
}

// ListarPecasInput é o contrato de entrada do caso de uso de listagem.
type ListarPecasInput struct {
	appquery.ParamsInput
}

// ListarPecasOutput é o contrato de saída do caso de uso de listagem.
type ListarPecasOutput struct {
	Items      []PecaOutput
	Total      int64
	Page       int
	PageSize   int
	TotalPages int
	Order      string
	Direction  string
}
