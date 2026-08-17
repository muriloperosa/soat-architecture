package cliente

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func novoClienteValido(t *testing.T) Cliente {
	t.Helper()

	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"  joÃO   da SILVA  ",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
	)

	require.NoError(t, err)

	return cliente
}

func TestNewClienteValido(t *testing.T) {
	antes := time.Now()

	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"  joÃO   da SILVA  ",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
	)

	depois := time.Now()

	require.NoError(t, err)

	require.Equal(t, uint64(0), cliente.ID())
	require.Equal(t, "52998224725", cliente.Documento().String())
	require.Equal(t, TipoPessoaFisica, cliente.Tipo())
	require.Equal(t, "João Da Silva", cliente.Nome())
	require.Equal(t, "joao@email.com", cliente.Email())
	require.Equal(t, "44999991234", cliente.Telefone().String())
	require.Equal(t, "senha123", cliente.Senha())
	require.True(t, cliente.Ativo())

	require.False(t, cliente.DataCadastro().IsZero())
	require.False(t, cliente.DataAtualizacao().IsZero())

	require.False(t, cliente.DataCadastro().Before(antes))
	require.False(t, cliente.DataCadastro().After(depois))

	require.Equal(t, cliente.DataCadastro(), cliente.DataAtualizacao())
}

func TestNewClienteNomeObrigatorio(t *testing.T) {
	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"   ",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
	)

	require.ErrorIs(t, err, ErrNomeObrigatorio)
	require.Equal(t, Cliente{}, cliente)
}

func TestNewClienteEmailObrigatorio(t *testing.T) {
	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"João da Silva",
		"",
		"(44) 99999-1234",
		"senha123",
	)

	require.ErrorIs(t, err, ErrEmailObrigatorio)
	require.Equal(t, Cliente{}, cliente)
}

func TestNewClienteSenhaObrigatoria(t *testing.T) {
	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"",
	)

	require.ErrorIs(t, err, ErrSenhaObrigatoria)
	require.Equal(t, Cliente{}, cliente)
}

func TestNewClienteDocumentoInvalido(t *testing.T) {
	cliente, err := NewCliente(
		"123.456.789-00",
		TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
	)

	require.ErrorIs(t, err, ErrCPFInvalido)
	require.Equal(t, Cliente{}, cliente)
}

func TestNewClienteTelefoneInvalido(t *testing.T) {
	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"João da Silva",
		"joao@email.com",
		"123",
		"senha123",
	)

	require.ErrorIs(t, err, ErrTelefoneInvalido)
	require.Equal(t, Cliente{}, cliente)
}

func TestClienteAtualizar(t *testing.T) {
	cliente := novoClienteValido(t)

	dataAtualizacaoAnterior := cliente.DataAtualizacao()

	time.Sleep(time.Millisecond)

	err := cliente.Atualizar("  mARIA   da SILVA ", "maria@email.com", "(44) 3031-1234")

	require.NoError(t, err)
	require.Equal(t, "Maria Da Silva", cliente.Nome())
	require.Equal(t, "maria@email.com", cliente.Email())
	require.Equal(t, "4430311234", cliente.Telefone().String())
	require.True(t, cliente.DataAtualizacao().After(dataAtualizacaoAnterior))
}

func TestClienteAtualizarNomeObrigatorio(t *testing.T) {
	cliente := novoClienteValido(t)

	nomeAnterior := cliente.Nome()
	emailAnterior := cliente.Email()
	telefoneAnterior := cliente.Telefone()
	dataAnterior := cliente.DataAtualizacao()

	err := cliente.Atualizar("","novo@email.com","(44) 3031-1234")

	require.ErrorIs(t, err, ErrNomeObrigatorio)

	require.Equal(t, nomeAnterior, cliente.Nome())
	require.Equal(t, emailAnterior, cliente.Email())
	require.Equal(t, telefoneAnterior, cliente.Telefone())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}

func TestClienteAtualizarEmailObrigatorio(t *testing.T) {
	cliente := novoClienteValido(t)

	nomeAnterior := cliente.Nome()
	emailAnterior := cliente.Email()
	telefoneAnterior := cliente.Telefone()
	dataAnterior := cliente.DataAtualizacao()

	err := cliente.Atualizar("Maria da Silva","","(44) 3031-1234")

	require.ErrorIs(t, err, ErrEmailObrigatorio)

	require.Equal(t, nomeAnterior, cliente.Nome())
	require.Equal(t, emailAnterior, cliente.Email())
	require.Equal(t, telefoneAnterior, cliente.Telefone())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}

func TestClienteAtualizarTelefoneInvalido(t *testing.T) {
	cliente := novoClienteValido(t)

	nomeAnterior := cliente.Nome()
	emailAnterior := cliente.Email()
	telefoneAnterior := cliente.Telefone()
	dataAnterior := cliente.DataAtualizacao()

	err := cliente.Atualizar("Maria da Silva","maria@email.com","123")

	require.ErrorIs(t, err, ErrTelefoneInvalido)

	require.Equal(t, nomeAnterior, cliente.Nome())
	require.Equal(t, emailAnterior, cliente.Email())
	require.Equal(t, telefoneAnterior, cliente.Telefone())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}

func TestClienteAlterarSenha(t *testing.T) {
	cliente := novoClienteValido(t)

	dataAnterior := cliente.DataAtualizacao()

	time.Sleep(time.Millisecond)

	err := cliente.AlterarSenha("novaSenha123")

	require.NoError(t, err)
	require.Equal(t, "novaSenha123", cliente.Senha())
	require.True(t, cliente.DataAtualizacao().After(dataAnterior))
}

func TestClienteAlterarSenhaObrigatoria(t *testing.T) {
	cliente := novoClienteValido(t)

	senhaAnterior := cliente.Senha()
	dataAnterior := cliente.DataAtualizacao()

	err := cliente.AlterarSenha("")

	require.ErrorIs(t, err, ErrSenhaObrigatoria)
	require.Equal(t, senhaAnterior, cliente.Senha())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}

func TestClienteInativar(t *testing.T) {
	cliente := novoClienteValido(t)

	dataAnterior := cliente.DataAtualizacao()

	time.Sleep(time.Millisecond)

	cliente.Inativar()

	require.False(t, cliente.Ativo())
	require.True(t, cliente.DataAtualizacao().After(dataAnterior))
}

func TestClienteInativarQuandoJaEstiverInativo(t *testing.T) {
	cliente := novoClienteValido(t)

	cliente.Inativar()
	dataAnterior := cliente.DataAtualizacao()

	cliente.Inativar()

	require.False(t, cliente.Ativo())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}

func TestClienteAtivar(t *testing.T) {
	cliente := novoClienteValido(t)

	cliente.Inativar()
	dataAnterior := cliente.DataAtualizacao()

	time.Sleep(time.Millisecond)

	cliente.Ativar()

	require.True(t, cliente.Ativo())
	require.True(t, cliente.DataAtualizacao().After(dataAnterior))
}

func TestClienteAtivarQuandoJaEstiverAtivo(t *testing.T) {
	cliente := novoClienteValido(t)

	dataAnterior := cliente.DataAtualizacao()

	cliente.Ativar()

	require.True(t, cliente.Ativo())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}
