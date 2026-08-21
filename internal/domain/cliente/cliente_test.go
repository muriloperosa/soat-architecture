package cliente

import (
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
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
		1,
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
		1,
	)

	depois := time.Now()

	require.NoError(t, err)

	require.Equal(t, uint64(0), cliente.ID())
	require.Equal(t, "52998224725", cliente.Documento().String())
	require.Equal(t, TipoPessoaFisica, cliente.Tipo())
	require.Equal(t, "João Da Silva", cliente.Nome())

	// Value Object Email
	require.Equal(t, "joao@email.com", cliente.Email().String())

	require.Equal(t, "44999991234", cliente.Telefone().String())

	// A senha armazenada é um SenhaHash, não a senha em texto puro.
	require.NotEmpty(t, cliente.Senha().String())
	require.NotEqual(t, "senha123", cliente.Senha().String())

	require.True(t, cliente.Ativo())
	require.True(t, cliente.RequerAlterarSenha())

	require.False(t, cliente.DataCadastro().IsZero())
	require.False(t, cliente.DataAtualizacao().IsZero())

	require.False(t, cliente.DataCadastro().Before(antes))
	require.False(t, cliente.DataCadastro().After(depois))

	require.Equal(
		t,
		cliente.DataCadastro(),
		cliente.DataAtualizacao(),
	)
}

func TestNewClienteNomeObrigatorio(t *testing.T) {
	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"   ",
		"joao@email.com",
		"(44) 99999-1234",
		"senha123",
		1,
	)

	require.ErrorIs(t, err, ErrNomeObrigatorio)
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
		1,
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
		1,
	)

	require.ErrorIs(t, err, ErrTelefoneInvalido)
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
		1,
	)

	require.Error(t, err)
	require.Equal(t, Cliente{}, cliente)
}

func TestNewClienteEmailInvalido(t *testing.T) {
	cliente, err := NewCliente(
		"529.982.247-25",
		TipoPessoaFisica,
		"João da Silva",
		"email-invalido",
		"(44) 99999-1234",
		"senha123",
		1,
	)

	require.Error(t, err)
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
		1,
	)

	require.Error(t, err)
	require.Equal(t, Cliente{}, cliente)
}

func TestClienteAtualizar(t *testing.T) {
	cliente := novoClienteValido(t)

	dataAnterior := cliente.DataAtualizacao()

	time.Sleep(time.Millisecond)

	err := cliente.Atualizar("  mARIA   da SILVA ", "MARIA@EMAIL.COM", "(44) 3031-1234")

	require.NoError(t, err)

	require.Equal(t, "Maria Da Silva", cliente.Nome())

	// O próprio VO normaliza o e-mail.
	require.Equal(t, "maria@email.com", cliente.Email().String())

	require.Equal(t, "4430311234", cliente.Telefone().String())

	require.True(t, cliente.DataAtualizacao().After(dataAnterior))
}

func TestClienteAtualizarNomeObrigatorio(t *testing.T) {
	cliente := novoClienteValido(t)

	nomeAnterior := cliente.Nome()
	emailAnterior := cliente.Email()
	telefoneAnterior := cliente.Telefone()
	dataAnterior := cliente.DataAtualizacao()

	err := cliente.Atualizar("", "novo@email.com", "(44) 3031-1234")

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

	err := cliente.Atualizar("Maria da Silva", "", "(44) 3031-1234")

	require.Error(t, err)

	// A entidade não deve ser alterada parcialmente.
	require.Equal(t, nomeAnterior, cliente.Nome())
	require.Equal(t, emailAnterior, cliente.Email())
	require.Equal(t, telefoneAnterior, cliente.Telefone())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}

func TestClienteAtualizarEmailInvalido(t *testing.T) {
	cliente := novoClienteValido(t)

	nomeAnterior := cliente.Nome()
	emailAnterior := cliente.Email()
	telefoneAnterior := cliente.Telefone()
	dataAnterior := cliente.DataAtualizacao()

	err := cliente.Atualizar("Maria da Silva", "email-invalido", "(44) 3031-1234")

	require.Error(t, err)

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

	err := cliente.Atualizar(
		"Maria da Silva",
		"maria@email.com",
		"123",
	)

	require.ErrorIs(t, err, ErrTelefoneInvalido)

	// Como a alteração somente ocorre após todas as validações,
	// nenhum dado deve ter sido modificado.
	require.Equal(t, nomeAnterior, cliente.Nome())
	require.Equal(t, emailAnterior, cliente.Email())
	require.Equal(t, telefoneAnterior, cliente.Telefone())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())
}

func TestClienteAlterarSenha(t *testing.T) {
	cliente := novoClienteValido(t)

	dataAnterior := cliente.DataAtualizacao()

	require.True(t, cliente.RequerAlterarSenha())
	require.True(t, cliente.Senha().Confere("senha123"))

	time.Sleep(time.Millisecond)

	err := cliente.AlterarSenha("novaSenha123")

	require.NoError(t, err)

	require.False(t, cliente.Senha().Confere("senha123"))
	require.True(t, cliente.Senha().Confere("novaSenha123"))
	require.False(t, cliente.RequerAlterarSenha())

	require.True(
		t,
		cliente.DataAtualizacao().After(dataAnterior),
	)
}

func TestClienteAlterarSenhaInvalida(t *testing.T) {
	cliente := novoClienteValido(t)

	senhaAnterior := cliente.Senha()
	dataAnterior := cliente.DataAtualizacao()

	require.True(t, cliente.RequerAlterarSenha())

	err := cliente.AlterarSenha("")

	require.Error(t, err)

	require.Equal(t, senhaAnterior, cliente.Senha())
	require.Equal(t, dataAnterior, cliente.DataAtualizacao())

	// Se a alteração falhou, a troca continua sendo obrigatória.
	require.True(t, cliente.RequerAlterarSenha())
}

func TestClienteDefinirID(t *testing.T) {
	cliente := novoClienteValido(t)
	require.Equal(t, uint64(0), cliente.ID())
	cliente.DefinirID(10)
	require.Equal(t, uint64(10), cliente.ID())
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

func TestRestaurarCliente(t *testing.T) {
	dataCadastro := time.Now().Add(-time.Hour)
	dataAtualizacao := time.Now()

	senhaHash, err := shared.NewSenhaHash("senha123")
	require.NoError(t, err)

	cliente, err := RestaurarCliente(
		10,
		"52998224725",
		TipoPessoaFisica,
		"João Da Silva",
		"joao@email.com",
		"44999991234",
		senhaHash.String(),
		1,
		false,
		true,
		dataCadastro,
		dataAtualizacao,
	)

	require.NoError(t, err)
	require.NotNil(t, cliente)

	require.Equal(t, uint64(10), cliente.ID())
	require.Equal(t, "52998224725", cliente.Documento().String())
	require.Equal(t, TipoPessoaFisica, cliente.Tipo())
	require.Equal(t, "João Da Silva", cliente.Nome())
	require.Equal(t, "joao@email.com", cliente.Email().String())
	require.Equal(t, "44999991234", cliente.Telefone().String())

	require.Equal(t, senhaHash.String(), cliente.Senha().String())
	require.True(t, cliente.Senha().Confere("senha123"))

	require.False(t, cliente.RequerAlterarSenha())
	require.True(t, cliente.Ativo())
	require.Equal(t, dataCadastro, cliente.DataCadastro())
	require.Equal(t, dataAtualizacao, cliente.DataAtualizacao())
}

func TestRestaurarClienteDocumentoInvalido(t *testing.T) {
	cliente, err := RestaurarCliente(
		10,
		"12345678900",
		TipoPessoaFisica,
		"João Da Silva",
		"joao@email.com",
		"44999991234",
		"senha123",
		1,
		true,
		true,
		time.Now(),
		time.Now(),
	)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, ErrCPFInvalido)
}

func TestRestaurarClienteTelefoneInvalido(t *testing.T) {
	cliente, err := RestaurarCliente(
		10,
		"52998224725",
		TipoPessoaFisica,
		"João Da Silva",
		"joao@email.com",
		"123",
		"senha123",
		1,
		true,
		true,
		time.Now(),
		time.Now(),
	)

	require.Nil(t, cliente)
	require.ErrorIs(t, err, ErrTelefoneInvalido)
}

func TestRestaurarClienteEmailInvalido(t *testing.T) {
	cliente, err := RestaurarCliente(
		10,
		"52998224725",
		TipoPessoaFisica,
		"João Da Silva",
		"email-invalido",
		"44999991234",
		"senha123",
		1,
		true,
		true,
		time.Now(),
		time.Now(),
	)

	require.Nil(t, cliente)
	require.Error(t, err)
}
