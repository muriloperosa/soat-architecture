package peca

import (
	"strconv"
	"time"
)

// Peca é o item de estoque da oficina que pode ser aplicado numa Ordem de
// Serviço. QuantidadeEmEstoque representa o estoque físico; reservas para
// ordens de serviço são controladas por outro agregado (ReservaPeca), não
// duplicadas aqui.
type Peca struct {
	id                  uint64
	codigo              string
	nome                string
	marca               string
	descricao           string
	preco               float64
	quantidadeEmEstoque int
	estoqueMinimo       int
	criadoPor           uint64
	ativo               bool
	dataCadastro        time.Time
	dataAtualizacao     time.Time
}

func NewPeca(nome, marca, descricao string, preco float64, quantidadeEmEstoque, estoqueMinimo int, criadoPor uint64) (*Peca, error) {
	if nome == "" {
		return nil, ErrNomeObrigatorio
	}

	if marca == "" {
		return nil, ErrMarcaObrigatoria
	}

	if descricao == "" {
		return nil, ErrDescricaoObrigatoria
	}

	if preco < 0 {
		return nil, ErrPrecoInvalido
	}

	if quantidadeEmEstoque < 0 {
		return nil, ErrQuantidadeEmEstoqueInvalida
	}

	if estoqueMinimo < 0 {
		return nil, ErrEstoqueMinimoInvalido
	}

	if criadoPor == 0 {
		return nil, ErrCriadoPorObrigatorio
	}

	agora := time.Now()

	return &Peca{
		codigo:              gerarCodigo(),
		nome:                nome,
		marca:               marca,
		descricao:           descricao,
		preco:               preco,
		quantidadeEmEstoque: quantidadeEmEstoque,
		estoqueMinimo:       estoqueMinimo,
		criadoPor:           criadoPor,
		ativo:               true,
		dataCadastro:        agora,
		dataAtualizacao:     agora,
	}, nil
}

// RestaurarPeca reidrata uma Peca a partir de dados já persistidos; não
// reaplica validação de negócio nem regera o código. Usado só pelo mapper
// de persistência (internal/infrastructure/persistence/mysql/peca).
func RestaurarPeca(id uint64, codigo, nome, marca, descricao string, preco float64, quantidadeEmEstoque, estoqueMinimo int, criadoPor uint64, ativo bool, dataCadastro, dataAtualizacao time.Time) *Peca {
	return &Peca{
		id:                  id,
		codigo:              codigo,
		nome:                nome,
		marca:               marca,
		descricao:           descricao,
		preco:               preco,
		quantidadeEmEstoque: quantidadeEmEstoque,
		estoqueMinimo:       estoqueMinimo,
		criadoPor:           criadoPor,
		ativo:               ativo,
		dataCadastro:        dataCadastro,
		dataAtualizacao:     dataAtualizacao,
	}
}

func gerarCodigo() string {
	return "P" + strconv.FormatInt(time.Now().UnixNano(), 36)
}

// Atualizar troca os dados cadastrais e o estoque mínimo. O estoque físico
// (QuantidadeEmEstoque) só muda via Consumir/Repor.
func (p *Peca) Atualizar(nome, marca, descricao string, preco float64, estoqueMinimo int) error {
	if nome == "" {
		return ErrNomeObrigatorio
	}

	if marca == "" {
		return ErrMarcaObrigatoria
	}

	if descricao == "" {
		return ErrDescricaoObrigatoria
	}

	if preco < 0 {
		return ErrPrecoInvalido
	}

	if estoqueMinimo < 0 {
		return ErrEstoqueMinimoInvalido
	}

	p.nome = nome
	p.marca = marca
	p.descricao = descricao
	p.preco = preco
	p.estoqueMinimo = estoqueMinimo
	p.dataAtualizacao = time.Now()

	return nil
}

// Consumir baixa quantidade do estoque físico. Não pode deixar o estoque
// abaixo do mínimo definido.
func (p *Peca) Consumir(quantidade int) error {
	if quantidade <= 0 {
		return ErrQuantidadeInvalida
	}

	if p.quantidadeEmEstoque-quantidade < p.estoqueMinimo {
		return ErrEstoqueInsuficiente
	}

	p.quantidadeEmEstoque -= quantidade
	p.dataAtualizacao = time.Now()

	return nil
}

// Repor adiciona quantidade ao estoque físico.
func (p *Peca) Repor(quantidade int) error {
	if quantidade <= 0 {
		return ErrQuantidadeInvalida
	}

	p.quantidadeEmEstoque += quantidade
	p.dataAtualizacao = time.Now()

	return nil
}

func (p *Peca) Ativar() {
	p.ativo = true
	p.dataAtualizacao = time.Now()
}

func (p *Peca) Inativar() {
	p.ativo = false
	p.dataAtualizacao = time.Now()
}

func (p *Peca) AtribuirID(id uint64) {
	p.id = id
}

func (p *Peca) ID() uint64                 { return p.id }
func (p *Peca) Codigo() string             { return p.codigo }
func (p *Peca) Nome() string               { return p.nome }
func (p *Peca) Marca() string              { return p.marca }
func (p *Peca) Descricao() string          { return p.descricao }
func (p *Peca) Preco() float64             { return p.preco }
func (p *Peca) QuantidadeEmEstoque() int   { return p.quantidadeEmEstoque }
func (p *Peca) EstoqueMinimo() int         { return p.estoqueMinimo }
func (p *Peca) CriadoPor() uint64          { return p.criadoPor }
func (p *Peca) Ativo() bool                { return p.ativo }
func (p *Peca) DataCadastro() time.Time    { return p.dataCadastro }
func (p *Peca) DataAtualizacao() time.Time { return p.dataAtualizacao }
