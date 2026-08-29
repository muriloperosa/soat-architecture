package peca

import (
	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
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

// ReservarPecaInput é o DTO de entrada do ReservarPecaUseCase.
type ReservarPecaInput struct {
	PecaID         uint64
	OrdemServicoID uint64
	Quantidade     int
}

// LiberarReservaPecaInput é o DTO de entrada do LiberarReservaPecaUseCase.
type LiberarReservaPecaInput struct {
	PecaID         uint64
	OrdemServicoID uint64
	Quantidade     int
}

// ReservaPecaOutput é o DTO de saída comum a ReservarPecaUseCase e
// LiberarReservaPecaUseCase.
type ReservaPecaOutput struct {
	ID             uint64
	OrdemServicoID uint64
	PecaID         uint64
	Quantidade     int
}

func toReservaOutput(r *domainreservapeca.ReservaPeca) ReservaPecaOutput {
	return ReservaPecaOutput{
		ID:             r.ID(),
		OrdemServicoID: r.OrdemServicoID(),
		PecaID:         r.PecaID(),
		Quantidade:     r.Quantidade(),
	}
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
