package cliente_test

import (
	"context"
	"errors"
	"testing"
	"time"

	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	mysqlcliente "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/cliente"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func clienteValido(t *testing.T) *domaincliente.Cliente {
	t.Helper()
	agora := time.Now()
	cliente, err := domaincliente.RestaurarCliente(
		1, "52998224725", domaincliente.TipoPessoaFisica, "João da Silva",
		"joao@email.com", "(44) 99999-1234", "senha123", 1, false, true, agora, agora,
	)
	require.NoError(t, err)
	return cliente
}

func TestCredenciaisAdapter_BuscarPorEmail_ComSucesso(t *testing.T) {
	repo := mocks.NewClienteRepository(t)
	cliente := clienteValido(t)
	repo.EXPECT().BuscarPorEmail(mock.Anything, "joao@email.com").Return(cliente, nil)

	adapter := mysqlcliente.NewCredenciaisAdapter(repo)
	credencial, err := adapter.BuscarPorEmail(context.Background(), "joao@email.com")

	require.NoError(t, err)
	require.Equal(t, uint64(1), credencial.ID)
	require.Equal(t, shared.PapelCliente, credencial.Papel)
	require.True(t, credencial.Ativo)
}

func TestCredenciaisAdapter_BuscarPorEmail_PropagaErro(t *testing.T) {
	repo := mocks.NewClienteRepository(t)
	erro := errors.New("conexao recusada")
	repo.EXPECT().BuscarPorEmail(mock.Anything, "joao@email.com").Return(nil, erro)

	adapter := mysqlcliente.NewCredenciaisAdapter(repo)
	credencial, err := adapter.BuscarPorEmail(context.Background(), "joao@email.com")

	require.Nil(t, credencial)
	require.ErrorIs(t, err, erro)
}

func TestCredenciaisAdapter_EstaAtivo_ComSucesso(t *testing.T) {
	repo := mocks.NewClienteRepository(t)
	cliente := clienteValido(t)
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(1)).Return(cliente, nil)

	adapter := mysqlcliente.NewCredenciaisAdapter(repo)
	ativo, err := adapter.EstaAtivo(context.Background(), 1)

	require.NoError(t, err)
	require.True(t, ativo)
}

func TestCredenciaisAdapter_EstaAtivo_PropagaErro(t *testing.T) {
	repo := mocks.NewClienteRepository(t)
	erro := errors.New("conexao recusada")
	repo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, erro)

	adapter := mysqlcliente.NewCredenciaisAdapter(repo)
	ativo, err := adapter.EstaAtivo(context.Background(), 999)

	require.False(t, ativo)
	require.ErrorIs(t, err, erro)
}
