package orcamento_test

import (
	"context"
	"errors"
	"testing"

	app "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	clientemocks "github.com/muriloperosa/soat-architecture/internal/domain/cliente/mocks"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainorcamentomocks "github.com/muriloperosa/soat-architecture/internal/domain/orcamento/mocks"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	ordemservicomocks "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico/mocks"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	sharedmocks "github.com/muriloperosa/soat-architecture/internal/domain/shared/mocks"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func clienteComEmail(t *testing.T, id uint64, email string) *domaincliente.Cliente {
	t.Helper()
	cliente, err := domaincliente.NewCliente("52998224725", domaincliente.TipoPessoaFisica, "Maria Silva", email, "11999998888", "senha123", 3)
	require.NoError(t, err)
	cliente.DefinirID(id)
	return &cliente
}

// osEmDiagnosticoComDiagnostico devolve uma OS EM_DIAGNOSTICO com diagnóstico
// preenchido, o estado exigido pra finalizar o orçamento. ordemServicoExistente
// já deixa a OS EM_DIAGNOSTICO (ver gerar_orcamento_test.go).
func osEmDiagnosticoComDiagnostico(t *testing.T, id uint64) *domainordemservico.OrdemServico {
	t.Helper()
	os := ordemServicoExistente(t, id)
	require.NoError(t, os.InformarDiagnostico("Falha na bomba de combustível"))
	return os
}

func TestFinalizarOrcamentoUseCase_ComSucesso(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	o := orcamentoVazio(t, 42)
	require.NoError(t, o.AdicionarItemServico(5, 2, 100.0, 60))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(o, nil)

	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	cliente := clienteComEmail(t, 10, "maria@email.com")
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(cliente, nil)

	var corpoCapturado string
	emailSender.EXPECT().
		Enviar(mock.Anything, "maria@email.com", mock.AnythingOfType("string"), mock.AnythingOfType("string")).
		Run(func(_ context.Context, _, _, corpo string) { corpoCapturado = corpo }).
		Return(nil)

	uc := app.NewFinalizarOrcamentoUseCase(orcamentoRepo, osRepo, clienteRepo, emailSender)
	output, err := uc.Executar(context.Background(), app.FinalizarOrcamentoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.NoError(t, err)
	require.Equal(t, 200.0, output.ValorTotal)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
	require.Contains(t, corpoCapturado, "Maria Silva")
	require.Contains(t, corpoCapturado, "R$ 200.00")
}

func TestFinalizarOrcamentoUseCase_OrcamentoInexistente(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)

	uc := app.NewFinalizarOrcamentoUseCase(orcamentoRepo, osRepo, clienteRepo, emailSender)
	_, err := uc.Executar(context.Background(), app.FinalizarOrcamentoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoNaoEncontrado)
}

func TestFinalizarOrcamentoUseCase_TransicaoInvalida(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	// OS ainda RECEBIDA (nunca passou por IniciarDiagnostico) não permite a
	// transição pra AGUARDANDO_APROVACAO.
	osRecebida, err := domainordemservico.NewOrdemServico("OS-20260827-a1b2c3d4e5f6", 10, 20, 52_300, "", "", 3)
	require.NoError(t, err)
	osRecebida.AtribuirID(42)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(osRecebida, nil)

	uc := app.NewFinalizarOrcamentoUseCase(orcamentoRepo, osRepo, clienteRepo, emailSender)
	_, err = uc.Executar(context.Background(), app.FinalizarOrcamentoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.ErrorIs(t, err, domainordemservico.ErrTransicaoStatusInvalida)
}

func TestFinalizarOrcamentoUseCase_ClienteInexistente(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(nil, domaincliente.ErrClienteNaoEncontrado)

	uc := app.NewFinalizarOrcamentoUseCase(orcamentoRepo, osRepo, clienteRepo, emailSender)
	_, err := uc.Executar(context.Background(), app.FinalizarOrcamentoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.ErrorIs(t, err, domaincliente.ErrClienteNaoEncontrado)
}

func TestFinalizarOrcamentoUseCase_ErroAoEnviarEmail(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)
	erroEnvio := errors.New("provedor indisponível")

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	cliente := clienteComEmail(t, 10, "maria@email.com")
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(cliente, nil)
	emailSender.EXPECT().
		Enviar(mock.Anything, mock.Anything, mock.Anything, mock.Anything).
		Return(erroEnvio)

	uc := app.NewFinalizarOrcamentoUseCase(orcamentoRepo, osRepo, clienteRepo, emailSender)
	_, err := uc.Executar(context.Background(), app.FinalizarOrcamentoInput{OrdemServicoID: 42, UsuarioID: 30})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, erroEnvio)
}
