package cliente_test

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/test/factory"
	"github.com/stretchr/testify/require"
)

func TestNewClienteValido(t *testing.T) {
	documento := factory.DocumentoPFValido(t)
	telefone := factory.TelefoneCelularValido(t)

	antes := time.Now()

	c, err := cliente.NewCliente(
		documento,
		cliente.TipoPessoaFisica,
		"  caio   henrique  ",
		"cliente@email.com",
		telefone,
		"Senha@123",
	)

	depois := time.Now()

	require.NoError(t, err)

	require.Zero(t, c.ID())
	require.Equal(t, documento, c.Documento())
	require.Equal(t, cliente.TipoPessoaFisica, c.Tipo())
	require.Equal(t, "Caio Henrique", c.Nome())
	require.Equal(t, "cliente@email.com", c.Email())
	require.Equal(t, telefone, c.Telefone())
	require.Equal(t, "Senha@123", c.Senha())
	require.True(t, c.Ativo())

	require.False(t, c.DataCadastro().Before(antes))
	require.False(t, c.DataCadastro().After(depois))
	require.False(t, c.DataAtualizacao().Before(antes))
	require.False(t, c.DataAtualizacao().After(depois))
}

func TestNewClienteNomeObrigatorio(t *testing.T) {
	documento := factory.DocumentoPFValido(t)
	telefone := factory.TelefoneCelularValido(t)

	c, err := cliente.NewCliente(
		documento,
		cliente.TipoPessoaFisica,
		"   ",
		"cliente@email.com",
		telefone,
		"Senha@123",
	)

	require.ErrorIs(t, err, cliente.ErrNomeObrigatorio)

	require.Zero(t, c)
}

func TestNewClienteEmailObrigatorio(t *testing.T) {
	documento := factory.DocumentoPFValido(t)
	telefone := factory.TelefoneCelularValido(t)

	c, err := cliente.NewCliente(
		documento,
		cliente.TipoPessoaFisica,
		"Caio Henrique",
		"",
		telefone,
		"Senha@123",
	)

	require.ErrorIs(t, err, cliente.ErrEmailObrigatorio)

	require.Zero(t, c)
}

func TestNewClienteSenhaObrigatoria(t *testing.T) {
	documento := factory.DocumentoPFValido(t)
	telefone := factory.TelefoneCelularValido(t)

	c, err := cliente.NewCliente(
		documento,
		cliente.TipoPessoaFisica,
		"Caio Henrique",
		"cliente@email.com",
		telefone,
		"",
	)

	require.ErrorIs(t, err, cliente.ErrSenhaObrigatoria)

	require.Zero(t, c)
}

func TestClienteAtualizar(t *testing.T) {
	c := factory.ClienteValido(t)
	telefone := factory.TelefoneFixoValido(t)

	antes := time.Now()
	err := c.Atualizar("Novo NOME", "novo@email.com", telefone)
	depois := time.Now()

	require.NoError(t, err)
	require.Equal(t, "Novo Nome", c.Nome())
	require.Equal(t, "novo@email.com", c.Email())
	require.Equal(t, telefone, c.Telefone())
	require.False(t, c.DataAtualizacao().Before(antes))
	require.False(t, c.DataAtualizacao().After(depois))
}

func TestClienteAtualizarNomeObrigatorio(t *testing.T) {
	c := factory.ClienteValido(t)

	nomeAnterior := c.Nome()
	emailAnterior := c.Email()
	telefoneAnterior := c.Telefone()
	dataAnterior := c.DataAtualizacao()

	err := c.Atualizar("", "novo@email.com", factory.TelefoneFixoValido(t))

	require.ErrorIs(t, err, cliente.ErrNomeObrigatorio)
	require.Equal(t, nomeAnterior, c.Nome())
	require.Equal(t, emailAnterior, c.Email())
	require.Equal(t, telefoneAnterior, c.Telefone())
	require.Equal(t, dataAnterior, c.DataAtualizacao())
}

func TestClienteAtualizarEmailObrigatorio(t *testing.T) {
	c := factory.ClienteValido(t)

	nomeAnterior := c.Nome()
	emailAnterior := c.Email()
	telefoneAnterior := c.Telefone()
	dataAnterior := c.DataAtualizacao()

	err := c.Atualizar("Novo Nome", "", factory.TelefoneFixoValido(t))

	require.ErrorIs(t, err, cliente.ErrEmailObrigatorio)
	require.Equal(t, nomeAnterior, c.Nome())
	require.Equal(t, emailAnterior, c.Email())
	require.Equal(t, telefoneAnterior, c.Telefone())
	require.Equal(t, dataAnterior, c.DataAtualizacao())
}

func TestClienteAlterarSenha(t *testing.T) {
	c := factory.ClienteValido(t)

	senhaAnterior := c.Senha()
	antes := time.Now()
	err := c.AlterarSenha("NovaSenha@123")
	depois := time.Now()

	require.NoError(t, err)

	require.NotEqual(t, senhaAnterior, c.Senha())
	require.Equal(t, "NovaSenha@123", c.Senha())
	require.False(t, c.DataAtualizacao().Before(antes))
	require.False(t, c.DataAtualizacao().After(depois))
}

func TestClienteAlterarSenhaObrigatoria(t *testing.T) {
	c := factory.ClienteValido(t)

	senhaAnterior := c.Senha()
	dataAnterior := c.DataAtualizacao()

	err := c.AlterarSenha("")

	require.ErrorIs(t, err, cliente.ErrSenhaObrigatoria)

	require.Equal(t, senhaAnterior, c.Senha())

	require.Equal(t, dataAnterior, c.DataAtualizacao())
}

func TestClienteInativar(t *testing.T) {
	c := factory.ClienteValido(t)

	require.True(t, c.Ativo())
	antes := time.Now()
	c.Inativar()
	depois := time.Now()

	require.False(t, c.Ativo())
	require.False(t, c.DataAtualizacao().Before(antes))

	require.False(t, c.DataAtualizacao().After(depois))
}

func TestClienteInativarQuandoJaInativoNaoAlteraData(t *testing.T) {
	c := factory.ClienteValido(t)

	c.Inativar()
	dataAnterior := c.DataAtualizacao()
	c.Inativar()

	require.False(t, c.Ativo())
	require.Equal(t, dataAnterior, c.DataAtualizacao())
}

func TestClienteAtivar(t *testing.T) {
	c := factory.ClienteValido(t)

	c.Inativar()
	require.False(t, c.Ativo())

	antes := time.Now()
	c.Ativar()
	depois := time.Now()

	require.True(t, c.Ativo())
	require.False(t, c.DataAtualizacao().Before(antes))
	require.False(t, c.DataAtualizacao().After(depois))
}

func TestClienteAtivarQuandoJaAtivoNaoAlteraData(t *testing.T) {
	c := factory.ClienteValido(t)

	dataAnterior := c.DataAtualizacao()
	c.Ativar()

	require.True(t, c.Ativo())
	require.Equal(t, dataAnterior, c.DataAtualizacao())
}
