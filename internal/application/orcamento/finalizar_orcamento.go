package orcamento

import (
	"context"
	"errors"
	"fmt"
	"strings"

	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	domainorcamento "github.com/muriloperosa/soat-architecture/internal/domain/orcamento"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

// FinalizarOrcamentoUseCase encerra a montagem do orçamento: move a OS de
// EM_DIAGNOSTICO para AGUARDANDO_APROVACAO e, como efeito da transição,
// envia o orçamento por e-mail ao cliente. Exige que a OS já tenha um
// orçamento gerado.
type FinalizarOrcamentoUseCase struct {
	orcamentoRepository    domainorcamento.OrcamentoRepository
	ordemServicoRepository domainordemservico.OrdemServicoRepository
	clienteRepository      domaincliente.ClienteRepository
	emailSender            shared.EmailSender
}

func NewFinalizarOrcamentoUseCase(
	orcamentoRepository domainorcamento.OrcamentoRepository,
	ordemServicoRepository domainordemservico.OrdemServicoRepository,
	clienteRepository domaincliente.ClienteRepository,
	emailSender shared.EmailSender,
) *FinalizarOrcamentoUseCase {
	return &FinalizarOrcamentoUseCase{
		orcamentoRepository:    orcamentoRepository,
		ordemServicoRepository: ordemServicoRepository,
		clienteRepository:      clienteRepository,
		emailSender:            emailSender,
	}
}

func (uc *FinalizarOrcamentoUseCase) Executar(ctx context.Context, input FinalizarOrcamentoInput) (OrcamentoOutput, error) {
	orcamento, err := uc.orcamentoRepository.BuscarPorOrdemServicoID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainorcamento.ErrOrcamentoNaoEncontrado) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar orçamento", err)
	}

	os, err := uc.ordemServicoRepository.BuscarPorID(ctx, input.OrdemServicoID)
	if err != nil {
		if errors.Is(err, domainordemservico.ErrOrdemServicoNaoEncontrada) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar ordem de serviço", err)
	}

	if err := os.EnviarParaAprovacao(input.UsuarioID); err != nil {
		return OrcamentoOutput{}, err
	}

	if err := uc.ordemServicoRepository.Atualizar(ctx, os); err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao atualizar ordem de serviço", err)
	}

	cliente, err := uc.clienteRepository.BuscarPorID(ctx, os.ClienteID())
	if err != nil {
		if errors.Is(err, domaincliente.ErrClienteNaoEncontrado) {
			return OrcamentoOutput{}, err
		}
		return OrcamentoOutput{}, shared.NewInternalError("erro ao buscar cliente", err)
	}

	assunto := fmt.Sprintf("Orçamento da Ordem de Serviço %s", os.Numero().String())
	corpo := montarCorpoEmail(os.Numero().String(), cliente.Nome(), orcamento)

	if err := uc.emailSender.Enviar(ctx, cliente.Email().String(), assunto, corpo); err != nil {
		return OrcamentoOutput{}, shared.NewInternalError("erro ao enviar e-mail do orçamento", err)
	}

	return toOutput(orcamento), nil
}

func montarCorpoEmail(numeroOS, nomeCliente string, o *domainorcamento.Orcamento) string {
	var b strings.Builder

	fmt.Fprintf(&b, "Olá, %s!\n\n", nomeCliente)
	fmt.Fprintf(&b, "Segue o orçamento da Ordem de Serviço %s:\n\n", numeroOS)

	if len(o.ItensServico()) > 0 {
		b.WriteString("Serviços:\n")
		for _, item := range o.ItensServico() {
			fmt.Fprintf(&b, "- Serviço #%d | quantidade: %d | valor unitário: R$ %.2f | subtotal: R$ %.2f\n",
				item.ServicoID(), item.Quantidade(), item.Valor(), item.CalcularSubtotal())
		}
		b.WriteString("\n")
	}

	if len(o.ItensPeca()) > 0 {
		b.WriteString("Peças:\n")
		for _, item := range o.ItensPeca() {
			fmt.Fprintf(&b, "- %s | quantidade: %d | valor unitário: R$ %.2f | subtotal: R$ %.2f\n",
				item.Descricao(), item.Quantidade(), item.Valor(), item.CalcularSubtotal())
		}
		b.WriteString("\n")
	}

	fmt.Fprintf(&b, "Valor total dos serviços: R$ %.2f\n", o.ValorItemServicos())
	fmt.Fprintf(&b, "Valor total das peças: R$ %.2f\n", o.ValorItemPecas())
	fmt.Fprintf(&b, "Valor total do orçamento: R$ %.2f\n", o.ValorTotal())

	if o.Observacoes() != "" {
		fmt.Fprintf(&b, "\nObservações: %s\n", o.Observacoes())
	}

	return b.String()
}
