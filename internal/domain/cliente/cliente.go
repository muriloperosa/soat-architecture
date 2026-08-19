package cliente

import (
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared/texts"
)

type Cliente struct {
	id                 uint64
	documento          Documento
	tipo               TipoPessoa
	nome               string
	email              shared.Email
	senha              shared.SenhaHash
	telefone           Telefone
	ativo              bool
	dataCadastro       time.Time
	dataAtualizacao    time.Time
	requerAlterarSenha bool
	criadoPor          uint64
}

func NewCliente(
	documento string,
	tipo TipoPessoa,
	nome string,
	email string, //Substitua pelo Value Object de email
	telefone string,
	senha string,
	criadoPor uint64,
) (Cliente, error) {
	nome = texts.NormalizeUcFirst(nome)

	if nome == "" {
		return Cliente{}, ErrNomeObrigatorio
	}
	if criadoPor == 0 {
		return Cliente{}, ErrCriadoPorObrigatorio
	}

	documentoVO, err := NewDocumento(documento, tipo)
	if err != nil {
		return Cliente{}, err
	}

	telefoneVO, err := NewTelefone(telefone)
	if err != nil {
		return Cliente{}, err
	}

	emailVO, err := shared.NewEmail(email)
	if err != nil {
		return Cliente{}, err
	}

	senhaVO, err := shared.NewSenhaHash(senha)
	if err != nil {
		return Cliente{}, err
	}

	agora := time.Now()

	return Cliente{
		documento:          documentoVO,
		tipo:               tipo,
		nome:               nome,
		email:              emailVO,
		telefone:           telefoneVO,
		senha:              senhaVO,
		ativo:              true,
		requerAlterarSenha: true,
		criadoPor:          criadoPor,
		dataCadastro:       agora,
		dataAtualizacao:    agora,
	}, nil
}

func (c *Cliente) Atualizar(nome string, email string, telefone string) error {
	nome = texts.NormalizeUcFirst(nome)

	if nome == "" {
		return ErrNomeObrigatorio
	}

	emailVO, err := shared.NewEmail(email)
	if err != nil {
		return err
	}

	telefoneVO, err := NewTelefone(telefone)
	if err != nil {
		return err
	}

	c.nome = nome
	c.email = emailVO
	c.telefone = telefoneVO
	c.dataAtualizacao = time.Now()

	return nil
}

func (c *Cliente) AlterarSenha(novaSenha string) error {
	senhaVO, err := shared.NewSenhaHash(novaSenha)
	if err != nil {
		return err
	}
	c.requerAlterarSenha = false
	c.senha = senhaVO
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

func (c *Cliente) DefinirID(id uint64) { c.id = id }

func (c Cliente) ID() uint64 { return c.id }

func (c Cliente) Documento() Documento { return c.documento }

func (c Cliente) Tipo() TipoPessoa { return c.tipo }

func (c Cliente) Nome() string { return c.nome }

func (c Cliente) Email() shared.Email { return c.email }

func (c Cliente) Telefone() Telefone { return c.telefone }

func (c Cliente) Senha() shared.SenhaHash { return c.senha }

func (c Cliente) Ativo() bool { return c.ativo }

func (c Cliente) DataCadastro() time.Time { return c.dataCadastro }

func (c Cliente) DataAtualizacao() time.Time { return c.dataAtualizacao }

func (c Cliente) RequerAlterarSenha() bool { return c.requerAlterarSenha }

func (c Cliente) CriadoPor() uint64 { return c.criadoPor }

func RestaurarCliente(
	id uint64,
	documento string,
	tipo TipoPessoa,
	nome string,
	email string,
	telefone string,
	senha string,
	criadoPor uint64,
	requerAlterarSenha bool,
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

	emailVO, err := shared.NewEmail(email)
	if err != nil {
		return nil, err
	}

	senhaVO := shared.RestaurarSenhaHash(senha)

	return &Cliente{
		id:                 id,
		documento:          documentoVO,
		tipo:               tipo,
		nome:               nome,
		email:              emailVO,
		senha:              senhaVO,
		requerAlterarSenha: requerAlterarSenha,
		criadoPor:          criadoPor,
		telefone:           telefoneVO,
		ativo:              ativo,
		dataCadastro:       dataCadastro,
		dataAtualizacao:    dataAtualizacao,
	}, nil
}
