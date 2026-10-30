package orcamento_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/application/notificacao"
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

func TestEnviarOrcamentoParaAprovacaoUseCase_ComSucesso(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	o, err := domainorcamento.NewOrcamento(42, "Inspecionar após o reparo", 3)
	require.NoError(t, err)
	require.NoError(t, o.AdicionarItemServico(5, 2, 100.0, 60))
	require.NoError(t, o.AdicionarItemPeca(8, "Filtro de óleo", 3, 50.0))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(o, nil)

	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)

	cliente := clienteComEmail(t, 10, "maria@email.com")
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(cliente, nil).Twice()

	var corpoOrcamento, corpoStatus string
	emailSender.EXPECT().
		Enviar(mock.Anything, "maria@email.com", mock.MatchedBy(func(assunto string) bool {
			return strings.Contains(assunto, "AGUARDANDO_APROVACAO")
		}), mock.AnythingOfType("string")).
		Run(func(_ context.Context, _, _, corpo string) { corpoStatus = corpo }).
		Return(nil)
	emailSender.EXPECT().
		Enviar(mock.Anything, "maria@email.com", mock.MatchedBy(func(assunto string) bool {
			return strings.Contains(assunto, "Orçamento da Ordem de Serviço")
		}), mock.AnythingOfType("string")).
		Run(func(_ context.Context, _, _, corpo string) { corpoOrcamento = corpo }).
		Return(nil)

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(
		orcamentoRepo,
		osRepo,
		notificacao.NewStatus(clienteRepo, emailSender),
		notificacao.NewOrcamento(clienteRepo, emailSender),
	)
	output, err := uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.NoError(t, err)
	require.Equal(t, 350.0, output.ValorTotal)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
	require.Contains(t, corpoStatus, "Maria Silva")
	require.Contains(t, corpoStatus, "Status anterior: EM_DIAGNOSTICO")
	require.Contains(t, corpoStatus, "Status atual: AGUARDANDO_APROVACAO")
	require.Contains(t, corpoOrcamento, "Serviço #5 | quantidade: 2 | valor unitário: R$ 100.00 | subtotal: R$ 200.00")
	require.Contains(t, corpoOrcamento, "Filtro de óleo | quantidade: 3 | valor unitário: R$ 50.00 | subtotal: R$ 150.00")
	require.Contains(t, corpoOrcamento, "Valor total dos serviços: R$ 200.00")
	require.Contains(t, corpoOrcamento, "Valor total das peças: R$ 150.00")
	require.Contains(t, corpoOrcamento, "Valor total do orçamento: R$ 350.00")
	require.Contains(t, corpoOrcamento, "Observações: Inspecionar após o reparo")
}

func TestEnviarOrcamentoParaAprovacaoUseCase_OrcamentoInexistente(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).
		Return(nil, domainorcamento.ErrOrcamentoNaoEncontrado)

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, notificacao.NewStatus(clienteRepo, emailSender), notificacao.NewOrcamento(clienteRepo, emailSender))
	_, err := uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoNaoEncontrado)
}

func TestEnviarOrcamentoParaAprovacaoUseCase_TransicaoInvalida(t *testing.T) {
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

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, notificacao.NewStatus(clienteRepo, emailSender), notificacao.NewOrcamento(clienteRepo, emailSender))
	_, err = uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.ErrorIs(t, err, domainordemservico.ErrTransicaoStatusInvalida)
}

func TestEnviarOrcamentoParaAprovacaoUseCase_ClienteInexistenteNaoDesfazStatus(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	o := orcamentoVazio(t, 42)
	require.NoError(t, o.AdicionarItemServico(5, 1, 100.0, 60))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(o, nil)
	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(nil, domaincliente.ErrClienteNaoEncontrado).Twice()

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, notificacao.NewStatus(clienteRepo, emailSender), notificacao.NewOrcamento(clienteRepo, emailSender))
	_, err := uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.NoError(t, err)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
}

func TestEnviarOrcamentoParaAprovacaoUseCase_FalhaDeEmailNaoDesfazStatus(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	o := orcamentoVazio(t, 42)
	require.NoError(t, o.AdicionarItemServico(5, 1, 100.0, 60))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(o, nil)
	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	cliente := clienteComEmail(t, 10, "maria@email.com")
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(cliente, nil).Twice()
	emailSender.EXPECT().
		Enviar(mock.Anything, mock.Anything, mock.MatchedBy(func(assunto string) bool {
			return strings.Contains(assunto, "AGUARDANDO_APROVACAO")
		}), mock.Anything).
		Return(errors.New("provedor indisponível"))
	emailSender.EXPECT().
		Enviar(mock.Anything, mock.Anything, mock.MatchedBy(func(assunto string) bool {
			return strings.Contains(assunto, "Orçamento da Ordem de Serviço")
		}), mock.Anything).
		Return(nil)

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, notificacao.NewStatus(clienteRepo, emailSender), notificacao.NewOrcamento(clienteRepo, emailSender))
	_, err := uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.NoError(t, err)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
}

func TestEnviarOrcamentoParaAprovacaoUseCase_FalhaDeEmailDoOrcamentoNaoDesfazStatus(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	o := orcamentoVazio(t, 42)
	require.NoError(t, o.AdicionarItemServico(5, 1, 100.0, 60))
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(o, nil)
	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)
	osRepo.EXPECT().Atualizar(mock.Anything, os).Return(nil)
	cliente := clienteComEmail(t, 10, "maria@email.com")
	clienteRepo.EXPECT().BuscarPorID(mock.Anything, uint64(10)).Return(cliente, nil).Twice()
	emailSender.EXPECT().
		Enviar(mock.Anything, mock.Anything, mock.MatchedBy(func(assunto string) bool {
			return strings.Contains(assunto, "AGUARDANDO_APROVACAO")
		}), mock.Anything).
		Return(nil)
	emailSender.EXPECT().
		Enviar(mock.Anything, mock.Anything, mock.MatchedBy(func(assunto string) bool {
			return strings.Contains(assunto, "Orçamento da Ordem de Serviço")
		}), mock.Anything).
		Return(errors.New("provedor indisponível"))

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, notificacao.NewStatus(clienteRepo, emailSender), notificacao.NewOrcamento(clienteRepo, emailSender))
	_, err := uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.NoError(t, err)
	require.Equal(t, domainordemservico.StatusAguardandoAprovacao, os.Status())
}

func TestEnviarOrcamentoParaAprovacaoUseCase_OrcamentoVazioNaoPodeSerEnviado(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	clienteRepo := clientemocks.NewClienteRepository(t)
	emailSender := sharedmocks.NewEmailSender(t)

	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	os := osEmDiagnosticoComDiagnostico(t, 42)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(os, nil)

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, notificacao.NewStatus(clienteRepo, emailSender), notificacao.NewOrcamento(clienteRepo, emailSender))
	_, err := uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	require.ErrorIs(t, err, domainorcamento.ErrOrcamentoVazio)
	require.Equal(t, domainordemservico.StatusEmDiagnostico, os.Status())
}

func TestEnviarOrcamentoParaAprovacaoUseCase_ErroAoConsultarOS(t *testing.T) {
	orcamentoRepo := domainorcamentomocks.NewOrcamentoRepository(t)
	osRepo := ordemservicomocks.NewOrdemServicoRepository(t)
	errConsulta := errors.New("banco indisponível")
	orcamentoRepo.EXPECT().BuscarPorOrdemServicoID(mock.Anything, uint64(42)).Return(orcamentoVazio(t, 42), nil)
	osRepo.EXPECT().BuscarPorID(mock.Anything, uint64(42)).Return(nil, errConsulta)

	uc := app.NewEnviarOrcamentoParaAprovacaoUseCase(orcamentoRepo, osRepo, nil, nil)
	_, err := uc.Executar(context.Background(), app.EnviarOrcamentoParaAprovacaoInput{OrdemServicoID: 42, UsuarioID: 30})

	var appErr *shared.AppError
	require.ErrorAs(t, err, &appErr)
	require.Equal(t, shared.KindInternal, appErr.Kind)
	require.ErrorIs(t, err, errConsulta)
}
