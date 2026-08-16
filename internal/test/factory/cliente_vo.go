package factory

import (
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/stretchr/testify/require"
)

func DocumentoPFValido(t testing.TB) cliente.Documento {
	t.Helper()
	documento, err := cliente.NewDocumento("529.982.247-25", cliente.TipoPessoaFisica)
	require.NoError(t, err)
	return documento
}

func DocumentoPJValido(t testing.TB) cliente.Documento {
	t.Helper()
	documento, err := cliente.NewDocumento("04.252.011/0001-10", cliente.TipoPessoaJuridica)
	require.NoError(t, err)
	return documento
}

func TelefoneCelularValido(t testing.TB) cliente.Telefone {
	t.Helper()
	telefone, err := cliente.NewTelefone("(44) 99999-1234")
	require.NoError(t, err)
	return telefone
}

func TelefoneFixoValido(t testing.TB) cliente.Telefone {
	t.Helper()
	telefone, err := cliente.NewTelefone("(44) 3031-1234")
	require.NoError(t, err)
	return telefone
}
