package factory

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/stretchr/testify/require"
)

func ClienteValido(t testing.TB) cliente.Cliente {
	t.Helper()

	c, err := cliente.NewCliente(
		DocumentoPFValido(t),
		cliente.TipoPessoaFisica,
		"Caio Henrique",
		"cliente@email.com",
		TelefoneCelularValido(t),
		"Senha@123",
	)

	require.NoError(t, err)

	return c
}
