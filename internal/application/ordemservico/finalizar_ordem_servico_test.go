package ordemservico_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	domain "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	pecamocks "github.com/muriloperosa/soat-architecture/internal/domain/peca/mocks"
	domainreservapeca "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"
	reservamocks "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/test/helpers"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestFinalizarOrdemServicoUseCase_ComSucessoConsomeReservas(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	runner := &helpers.TransactionRunnerMock{}
	os := ordemServicoEmExecucao(t)
	peca := pecaComEstoqueFinalizacao(t, 1, 10, 2)
	reserva := reservaDePeca(t, 42, 1, 3)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(42)).Return(os, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return([]*domainreservapeca.ReservaPeca{reserva}, nil)
	pecaRepo.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(peca, nil)
	pecaRepo.EXPECT().Atualizar(mock.Anything, peca).
		Run(func(_ context.Context, atualizada *domainpeca.Peca) {
			require.Equal(t, 7, atualizada.QuantidadeEmEstoque())
		}).
		Return(nil)
	reservaRepo.EXPECT().Remover(mock.Anything, uint64(42), uint64(1)).Return(nil)
	repository.EXPECT().Atualizar(mock.Anything, os).
		Run(func(_ context.Context, atualizada *domain.OrdemServico) {
			require.Equal(t, domain.StatusFinalizada, atualizada.Status())
			require.Equal(t, domain.StatusFinalizada, atualizada.HistoricoStatus()[len(atualizada.HistoricoStatus())-1].Status())
			require.Equal(t, uint64(7), atualizada.HistoricoStatus()[len(atualizada.HistoricoStatus())-1].AlteradoPor())
		}).
		Return(nil)

	uc := app.NewFinalizarOrdemServicoUseCase(repository, pecaRepo, reservaRepo, runner)
	output, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{
		OrdemServicoID: 42,
		UsuarioID:      7,
	})

	require.NoError(t, err)
	require.Equal(t, uint64(42), output.ID)
	require.Equal(t, domain.StatusFinalizada.String(), output.Status)
	require.Equal(t, 1, runner.Calls)
}

func TestFinalizarOrdemServicoUseCase_SemReservasAindaFinaliza(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	os := ordemServicoEmExecucao(t)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(42)).Return(os, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return(nil, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	uc := app.NewFinalizarOrdemServicoUseCase(repository, pecaRepo, reservaRepo, &helpers.TransactionRunnerMock{})
	output, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	require.NoError(t, err)
	require.Equal(t, domain.StatusFinalizada.String(), output.Status)
}

func TestFinalizarOrdemServicoUseCase_OSInexistente(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(999)).Return(nil, domain.ErrOrdemServicoNaoEncontrada)

	uc := app.NewFinalizarOrdemServicoUseCase(
		repository,
		pecamocks.NewRepository(t),
		reservamocks.NewRepository(t),
		&helpers.TransactionRunnerMock{},
	)
	_, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{OrdemServicoID: 999, UsuarioID: 7})

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}

func TestFinalizarOrdemServicoUseCase_ErroAoBuscar(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	erroBanco := errors.New("banco indisponível")
	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(42)).Return(nil, erroBanco)

	uc := app.NewFinalizarOrdemServicoUseCase(
		repository,
		pecamocks.NewRepository(t),
		reservamocks.NewRepository(t),
		&helpers.TransactionRunnerMock{},
	)
	_, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}

func TestFinalizarOrdemServicoUseCase_OSNula(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(42)).Return(nil, nil)

	uc := app.NewFinalizarOrdemServicoUseCase(
		repository,
		pecamocks.NewRepository(t),
		reservamocks.NewRepository(t),
		&helpers.TransactionRunnerMock{},
	)
	_, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	require.ErrorIs(t, err, domain.ErrOrdemServicoNaoEncontrada)
}

func TestFinalizarOrdemServicoUseCase_TransicaoInvalida(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	os := ordemServicoRecebida(t)
	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(42)).Return(os, nil)

	uc := app.NewFinalizarOrdemServicoUseCase(
		repository,
		pecamocks.NewRepository(t),
		reservamocks.NewRepository(t),
		&helpers.TransactionRunnerMock{},
	)
	_, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	require.ErrorIs(t, err, domain.ErrTransicaoStatusInvalida)
}

func TestFinalizarOrdemServicoUseCase_EstoqueAbaixoDoMinimo(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	pecaRepo := pecamocks.NewRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	os := ordemServicoEmExecucao(t)
	peca := pecaComEstoqueFinalizacao(t, 1, 5, 4)
	reserva := reservaDePeca(t, 42, 1, 3)

	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(42)).Return(os, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return([]*domainreservapeca.ReservaPeca{reserva}, nil)
	pecaRepo.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(1)).Return(peca, nil)

	uc := app.NewFinalizarOrdemServicoUseCase(repository, pecaRepo, reservaRepo, &helpers.TransactionRunnerMock{})
	_, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	require.ErrorIs(t, err, domainpeca.ErrEstoqueInsuficiente)
	require.Equal(t, domain.StatusEmExecucao, os.Status())
}

func TestFinalizarOrdemServicoUseCase_ErroAoPersistir(t *testing.T) {
	repository := ordemservicomocks.NewOrdemServicoRepository(t)
	reservaRepo := reservamocks.NewRepository(t)
	os := ordemServicoEmExecucao(t)
	erroBanco := errors.New("banco indisponível")
	repository.EXPECT().BuscarPorIDComBloqueio(mock.Anything, uint64(42)).Return(os, nil)
	reservaRepo.EXPECT().BuscarPorOrdemServico(mock.Anything, uint64(42)).Return(nil, nil)
	repository.EXPECT().Atualizar(mock.Anything, os).Return(erroBanco)

	uc := app.NewFinalizarOrdemServicoUseCase(
		repository,
		pecamocks.NewRepository(t),
		reservaRepo,
		&helpers.TransactionRunnerMock{},
	)
	_, err := uc.Executar(context.Background(), app.FinalizarOrdemServicoInput{OrdemServicoID: 42, UsuarioID: 7})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroBanco)
}

func ordemServicoEmExecucao(t *testing.T) *domain.OrdemServico {
	t.Helper()
	os := ordemServicoAprovada(t)
	require.NoError(t, os.IniciarExecucao(3))
	return os
}

func pecaComEstoqueFinalizacao(t *testing.T, id uint64, estoque, minimo int) *domainpeca.Peca {
	t.Helper()
	p, err := domainpeca.NewPeca("Pastilha", "Bosch", "Dianteira", 89.9, estoque, minimo, 1)
	require.NoError(t, err)
	p.AtribuirID(id)
	return p
}

func reservaDePeca(t *testing.T, ordemServicoID, pecaID uint64, quantidade int) *domainreservapeca.ReservaPeca {
	t.Helper()
	reserva, err := domainreservapeca.NewReservaPeca(ordemServicoID, pecaID, quantidade)
	require.NoError(t, err)
	reserva.AtribuirID(99)
	return reserva
}
