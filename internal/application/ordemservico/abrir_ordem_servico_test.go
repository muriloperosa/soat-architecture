package ordemservico_test

import (
	"context"
	"errors"
	"regexp"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	domainveiculo "github.com/muriloperosa/soat-architecture/internal/domain/veiculo"
	veiculomocks "github.com/muriloperosa/soat-architecture/internal/domain/veiculo/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestAbrirOrdemServicoUseCase_ComSucesso(t *testing.T) {
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)

	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(clienteValido(t), nil)
	veiculoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(20)).Return(veiculoValido(t), nil)
	ordemRepo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*ordemservico.OrdemServico")).
		Run(func(_ context.Context, os *domainordemservico.OrdemServico) {
			require.Equal(t, domainordemservico.StatusRecebida, os.Status())
			require.Len(t, os.HistoricoStatus(), 1)
			require.Equal(t, domainordemservico.StatusRecebida, os.HistoricoStatus()[0].Status())
			os.AtribuirID(99)
		}).
		Return(nil)

	uc := app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo)
	out, err := uc.Executar(context.Background(), app.AbrirOrdemServicoInput{
		ClienteID:            10,
		VeiculoID:            20,
		QuilometragemEntrada: 52_300,
		Observacoes:          "Cliente relatou ruído no motor",
		UsuarioID:            30,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(99), out.ID)
	require.Regexp(t, regexp.MustCompile(`^OS-\d{8}-[0-9a-f]{12}$`), out.Numero)
	require.Equal(t, uint64(10), out.ClienteID)
	require.Equal(t, uint64(20), out.VeiculoID)
	require.Equal(t, uint32(52_300), out.QuilometragemEntrada)
	require.Equal(t, domainordemservico.StatusRecebida.String(), out.Status)
	require.Empty(t, out.Diagnostico)
	require.Equal(t, "Cliente relatou ruído no motor", out.Observacoes)
}

func TestAbrirOrdemServicoUseCase_ClienteInexistente(t *testing.T) {
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domaincliente.ErrClienteNaoEncontrado)

	uc := app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo)
	out, err := uc.Executar(context.Background(), app.AbrirOrdemServicoInput{ClienteID: 999, VeiculoID: 20, UsuarioID: 30})

	require.Empty(t, out)
	require.ErrorIs(t, err, domaincliente.ErrClienteNaoEncontrado)
}

func TestAbrirOrdemServicoUseCase_VeiculoInexistente(t *testing.T) {
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(clienteValido(t), nil)
	veiculoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(999)).Return(nil, domainveiculo.ErrVeiculoNaoEncontrado)

	uc := app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo)
	out, err := uc.Executar(context.Background(), app.AbrirOrdemServicoInput{ClienteID: 10, VeiculoID: 999, UsuarioID: 30})

	require.Empty(t, out)
	require.ErrorIs(t, err, domainveiculo.ErrVeiculoNaoEncontrado)
}

func TestAbrirOrdemServicoUseCase_ClienteRetornadoNulo(t *testing.T) {
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(nil, nil)

	uc := app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo)
	out, err := uc.Executar(context.Background(), app.AbrirOrdemServicoInput{ClienteID: 10, VeiculoID: 20, UsuarioID: 30})

	require.Empty(t, out)
	require.ErrorIs(t, err, domaincliente.ErrClienteNaoEncontrado)
}

func TestAbrirOrdemServicoUseCase_VeiculoRetornadoNulo(t *testing.T) {
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(clienteValido(t), nil)
	veiculoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(20)).Return(nil, nil)

	uc := app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo)
	out, err := uc.Executar(context.Background(), app.AbrirOrdemServicoInput{ClienteID: 10, VeiculoID: 20, UsuarioID: 30})

	require.Empty(t, out)
	require.ErrorIs(t, err, domainveiculo.ErrVeiculoNaoEncontrado)
}

func TestAbrirOrdemServicoUseCase_DadosInvalidos(t *testing.T) {
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(clienteValido(t), nil)
	veiculoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(20)).Return(veiculoValido(t), nil)

	uc := app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo)
	out, err := uc.Executar(context.Background(), app.AbrirOrdemServicoInput{ClienteID: 10, VeiculoID: 20, UsuarioID: 0})

	require.Empty(t, out)
	require.ErrorIs(t, err, domainordemservico.ErrCriadoPorObrigatorio)
}

func TestAbrirOrdemServicoUseCase_ErroAoPersistir(t *testing.T) {
	ordemRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	veiculoRepo := veiculomocks.NewRepository(t)
	erroBanco := errors.New("banco indisponível")

	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(clienteValido(t), nil)
	veiculoRepo.EXPECT().BuscarPorID(mock.Anything, uint64(20)).Return(veiculoValido(t), nil)
	ordemRepo.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*ordemservico.OrdemServico")).Return(erroBanco)

	uc := app.NewAbrirOrdemServicoUseCase(ordemRepo, clienteRepo, veiculoRepo)
	_, err := uc.Executar(context.Background(), app.AbrirOrdemServicoInput{ClienteID: 10, VeiculoID: 20, UsuarioID: 30})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}

func clienteValido(t *testing.T) *domaincliente.Cliente {
	t.Helper()
	cliente, err := domaincliente.NewCliente(
		"52998224725",
		domaincliente.TipoPessoaFisica,
		"Maria Silva",
		"maria@email.com",
		"11999998888",
		"senha123",
		30,
	)
	require.NoError(t, err)
	cliente.DefinirID(10)
	return &cliente
}

func veiculoValido(t *testing.T) *domainveiculo.Veiculo {
	t.Helper()
	veiculo, err := domainveiculo.NewVeiculo("ABC1D23", "Fiat", "Uno", 52_000, 2020, "Prata", 30)
	require.NoError(t, err)
	veiculo.AtribuirID(20)
	return veiculo
}
