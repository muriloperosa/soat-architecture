package peca

import (
	"context"
	"errors"
	"testing"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func pecaComEstoque(t *testing.T, estoque, estoqueMinimo int) *domain.Peca {
	t.Helper()

	p, err := domain.NewPeca("Peca 1", "Marca 1", "Descricao 1", 100.0, estoque, estoqueMinimo, 1)
	require.NoError(t, err)
	p.AtribuirID(1)

	return p
}

func TestNewReservarPecaUseCase(t *testing.T) {
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)

	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	require.NotNil(t, useCase)
	require.Equal(t, repository, useCase.repository)
	require.Equal(t, reservaRepository, useCase.reservaRepository)
}

func TestReservarPecaUseCaseExecutar_CriaReservaNova(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	useCase := NewReservarPecaUseCase(repository, reservaRepository, runner)

	p := pecaComEstoque(t, 10, 5)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(2, nil).Once()
	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(mock.Anything, uint64(10), uint64(1)).Return(nil, domainreservapeca.ErrReservaNaoEncontrada).Once()
	reservaRepository.EXPECT().
		Salvar(mock.Anything, mock.AnythingOfType("*reservapeca.ReservaPeca")).
		Run(func(ctx context.Context, r *domainreservapeca.ReservaPeca) { r.AtribuirID(99) }).
		Return(nil).
		Once()

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 3})

	require.NoError(t, err)
	require.Equal(t, uint64(99), output.ID)
	require.Equal(t, uint64(10), output.OrdemServicoID)
	require.Equal(t, uint64(1), output.PecaID)
	require.Equal(t, 3, output.Quantidade)
	require.Equal(t, 1, runner.Calls, "use case deveria delegar ao TransactionRunner, não rodar a lógica direto")
}

func TestReservarPecaUseCaseExecutar_IncrementaReservaExistente(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	p := pecaComEstoque(t, 10, 0)

	existente, err := domainreservapeca.NewReservaPeca(10, 1, 2)
	require.NoError(t, err)
	existente.AtribuirID(5)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(2, nil).Once()
	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(mock.Anything, uint64(10), uint64(1)).Return(existente, nil).Once()
	reservaRepository.EXPECT().Atualizar(mock.Anything, existente).Return(nil).Once()

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 3})

	require.NoError(t, err)
	require.Equal(t, uint64(5), output.ID)
	require.Equal(t, 5, output.Quantidade)
	require.Equal(t, 5, existente.Quantidade())
}

func TestReservarPecaUseCaseExecutar_PecaNaoEncontrada_RetornaErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(999)).Return(nil, domain.ErrPecaNaoEncontrada).Once()

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 999, OrdemServicoID: 10, Quantidade: 3})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrPecaNaoEncontrada)
}

func TestReservarPecaUseCaseExecutar_ErroAoSomarReservado_RetornaErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	p := pecaComEstoque(t, 10, 5)
	erroBanco := errors.New("erro ao somar reservas")

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(0, erroBanco).Once()

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 3})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, erroBanco)
}

func TestReservarPecaUseCaseExecutar_QuantidadeIndisponivel_RetornaErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	// estoque 10, minimo 5, reservado 2: disponivel pra reservar sem furar o minimo é 3 (10-2-5=3)
	p := pecaComEstoque(t, 10, 5)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(2, nil).Once()

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 8})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, domain.ErrQuantidadeIndisponivelParaReserva)
}

func TestReservarPecaUseCaseExecutar_QuantidadeZeroOuNegativa_RetornaErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 0})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, domainreservapeca.ErrQuantidadeInvalida)
}

func TestReservarPecaUseCaseExecutar_ErroAoBuscarReservaExistente_RetornaErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	p := pecaComEstoque(t, 10, 5)
	erroBanco := errors.New("erro ao consultar reserva")

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(2, nil).Once()
	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(mock.Anything, uint64(10), uint64(1)).Return(nil, erroBanco).Once()

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 3})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, erroBanco)
}

func TestReservarPecaUseCaseExecutar_ErroAoSalvar_RetornaErro(t *testing.T) {
	ctx := context.Background()
	repository := mocks.NewRepository(t)
	reservaRepository := reservamocks.NewRepository(t)
	useCase := NewReservarPecaUseCase(repository, reservaRepository, &helpers.TransactionRunnerMock{})

	p := pecaComEstoque(t, 10, 5)
	erroBanco := errors.New("erro ao salvar reserva")

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(p, nil).Once()
	reservaRepository.EXPECT().SomarQuantidadeReservada(mock.Anything, uint64(1)).Return(2, nil).Once()
	reservaRepository.EXPECT().BuscarPorOrdemEPecaComBloqueio(mock.Anything, uint64(10), uint64(1)).Return(nil, domainreservapeca.ErrReservaNaoEncontrada).Once()
	reservaRepository.EXPECT().Salvar(mock.Anything, mock.AnythingOfType("*reservapeca.ReservaPeca")).Return(erroBanco).Once()

	output, err := useCase.Executar(ctx, ReservarPecaInput{PecaID: 1, OrdemServicoID: 10, Quantidade: 3})

	require.Equal(t, ReservaPecaOutput{}, output)
	require.ErrorIs(t, err, erroBanco)
}
