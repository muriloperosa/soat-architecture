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
	documento string,
	tipo TipoPessoa,
	nome string,
	email string, //Substitua pelo Value Object de email
	telefone string,
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

	documentoVO, err := NewDocumento(documento, tipo)
	if err != nil {
		return Cliente{}, err
	}

	telefoneVO, err := NewTelefone(telefone)
	if err != nil {
		return Cliente{}, err
	}

	agora := time.Now()

	return Cliente{
		documento:       documentoVO,
		tipo:            tipo,
		nome:            nome,
		email:           email,
		telefone:        telefoneVO,
		senha:           senha,
		ativo:           true,
		dataCadastro:    agora,
		dataAtualizacao: agora,
	}, nil
}

func (c *Cliente) Atualizar(nome string, email string, telefone string) error {
	nome = texts.NormalizeUcFirst(nome)

	if nome == "" {
		return ErrNomeObrigatorio
	}

	if email == "" {
		return ErrEmailObrigatorio
	}

	telefoneVO, err := NewTelefone(telefone)
	if err != nil {
		return err
	}

	c.nome = nome
	c.email = email
	c.telefone = telefoneVO
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

func (c *Cliente) DefinirID(ID uint64) { c.id = ID }

func (c Cliente) ID() uint64 { return c.id }

func (c Cliente) Documento() Documento { return c.documento }

func (c Cliente) Tipo() TipoPessoa { return c.tipo }

func (c Cliente) Nome() string { return c.nome }

func (c Cliente) Email() string { return c.email }

func (c Cliente) Telefone() Telefone { return c.telefone }

func (c Cliente) Senha() string { return c.senha }

func (c Cliente) Ativo() bool { return c.ativo }

func (c Cliente) DataCadastro() time.Time { return c.dataCadastro }

func (c Cliente) DataAtualizacao() time.Time { return c.dataAtualizacao }

func ReidratarCliente(
	id uint64,
	documento string,
	tipo TipoPessoa,
	nome string,
	email string,
	telefone string,
	senha string,
	ativo bool,
	dataCadastro time.Time,
	dataAtualizacao time.Time,
) (*Cliente, error) {
	documentoVO, err := NewDocumento(documento, tipo)
	if err != nil {
		return nil, err
	}

	telefoneVO, err := NewTelefone(telefone)
	if err != nil {
		return nil, err
	}

	return &Cliente{
		id:              id,
		documento:       documentoVO,
		tipo:            tipo,
		nome:            nome,
		email:           email,
		senha:           senha,
		telefone:        telefoneVO,
		ativo:           ativo,
		dataCadastro:    dataCadastro,
		dataAtualizacao: dataAtualizacao,
	}, nil
}
