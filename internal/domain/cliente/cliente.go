package cliente

import (
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

type Cliente struct {
	id              uint64
	documento       Documento
	tipo            TipoPessoa
	nome            string
	email           string //Substitua pelo Value Object de email
	senha           string //Substitua pelo Value Object de senha
	telefone        Telefone
	ativo           bool
	dataCadastro    time.Time
	dataAtualizacao time.Time
}

func NewCliente(
	documento Documento,
	tipo TipoPessoa,
	nome string,
	email string, //Substitua pelo Value Object de email
	telefone Telefone,
	senha string, //Substitua pelo Value Object de senha
) (Cliente, error) {
	nome = texts.NormalizeUcFirst(nome)

	if nome == "" {
		return Cliente{}, ErrNomeObrigatorio
	}

	if email == "" { //Substitua pelo Value Object de email
		return Cliente{}, ErrEmailObrigatorio
	}

	if senha == "" { //Substitua pelo Value Object de senha
		return Cliente{}, ErrSenhaObrigatoria
	}

	return Cliente{
		documento:       documento,
		tipo:            tipo,
		nome:            nome,
		email:           email,
		telefone:        telefone,
		senha:           senha,
		ativo:           true,
		dataCadastro:    time.Now(),
		dataAtualizacao: time.Now(),
	}, nil
}

func (c *Cliente) Atualizar(nome string, email string, telefone Telefone) error {
	if nome == "" {
		return ErrNomeObrigatorio
	}

	if email == "" { //Substitua pelo Value Object de email
		return ErrEmailObrigatorio
	}

	c.nome = texts.NormalizeUcFirst(nome)
	c.email = email
	c.telefone = telefone
	c.dataAtualizacao = time.Now()
	return nil
}

func (c *Cliente) AlterarSenha(novaSenha string) error {
	if novaSenha == "" {
		return ErrSenhaObrigatoria
	}

	c.senha = novaSenha
	c.dataAtualizacao = time.Now()

	return nil
}

func (c *Cliente) Ativar() {
	if c.ativo {
		return
	}

	c.ativo = true
	c.dataAtualizacao = time.Now()
}

func (c *Cliente) Inativar() {
	if !c.ativo {
		return
	}

	c.ativo = false
	c.dataAtualizacao = time.Now()
}

func (c Cliente) ID() uint64 {
	return c.id
}

func (c Cliente) Documento() Documento {
	return c.documento
}

func (c Cliente) Tipo() TipoPessoa {
	return c.tipo
}

func (c Cliente) Nome() string {
	return c.nome
}

func (c Cliente) Email() string { //Substitua pelo Value Object de email
	return c.email
}

func (c Cliente) Telefone() Telefone {
	return c.telefone
}

func (c Cliente) Senha() string { //Substitua pelo Value Object de senha
	return c.senha
}

func (c Cliente) Ativo() bool {
	return c.ativo
}

func (c Cliente) DataCadastro() time.Time {
	return c.dataCadastro
}

func (c Cliente) DataAtualizacao() time.Time {
	return c.dataAtualizacao
}
