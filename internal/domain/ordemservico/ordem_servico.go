// Package ordemservico contém o Aggregate Root OrdemServico e seus invariantes.
package ordemservico

import (
	"math"
	"strings"
	"time"
)

// OrdemServico representa um atendimento realizado pela oficina.
type OrdemServico struct {
	id                   uint64
	numero               NumeroOrdemServico
	clienteID            uint64
	veiculoID            uint64
	quilometragemEntrada int
	status               StatusOrdemServico
	diagnostico          string
	observacoes          string
	criadoPor            uint64
	historicoStatus      []HistoricoStatus
	dataCadastro         time.Time
	dataAtualizacao      time.Time
}

// NewOrdemServico abre uma OS no estado RECEBIDA e registra seu primeiro histórico.
func NewOrdemServico(
	numero string,
	clienteID, veiculoID uint64,
	quilometragemEntrada int,
	diagnostico, observacoes string,
	criadoPor uint64,
) (*OrdemServico, error) {
	if clienteID == 0 {
		return nil, ErrClienteObrigatorio
	}
	if veiculoID == 0 {
		return nil, ErrVeiculoObrigatorio
	}
	if quilometragemEntrada < 0 || uint64(quilometragemEntrada) > math.MaxUint32 {
		return nil, ErrQuilometragemEntradaInvalida
	}
	if criadoPor == 0 {
		return nil, ErrCriadoPorObrigatorio
	}

	numeroVO, err := NewNumeroOrdemServico(numero)
	if err != nil {
		return nil, err
	}

	agora := time.Now()

	historicoInicial, err := NewHistoricoStatus(StatusRecebida, criadoPor, "", agora)
	if err != nil {
		return nil, err
	}

	return &OrdemServico{
		numero:               numeroVO,
		clienteID:            clienteID,
		veiculoID:            veiculoID,
		quilometragemEntrada: quilometragemEntrada,
		status:               StatusRecebida,
		diagnostico:          strings.TrimSpace(diagnostico),
		observacoes:          strings.TrimSpace(observacoes),
		criadoPor:            criadoPor,
		historicoStatus:      []HistoricoStatus{historicoInicial},
		dataCadastro:         agora,
		dataAtualizacao:      agora,
	}, nil
}

// ReidratarOrdemServico recompõe uma OS a partir dos dados persistidos.
func ReidratarOrdemServico(
	id uint64,
	numero NumeroOrdemServico,
	clienteID, veiculoID uint64,
	quilometragemEntrada int,
	status StatusOrdemServico,
	diagnostico, observacoes string,
	criadoPor uint64,
	historicoStatus []HistoricoStatus,
	dataCadastro, dataAtualizacao time.Time,
) *OrdemServico {
	historico := append([]HistoricoStatus(nil), historicoStatus...)

	return &OrdemServico{
		id:                   id,
		numero:               numero,
		clienteID:            clienteID,
		veiculoID:            veiculoID,
		quilometragemEntrada: quilometragemEntrada,
		status:               status,
		diagnostico:          diagnostico,
		observacoes:          observacoes,
		criadoPor:            criadoPor,
		historicoStatus:      historico,
		dataCadastro:         dataCadastro,
		dataAtualizacao:      dataAtualizacao,
	}
}

// AtribuirID registra a identidade gerada pela persistência na OS e no histórico inicial.
func (o *OrdemServico) AtribuirID(id uint64) {
	o.id = id
	for i := range o.historicoStatus {
		if o.historicoStatus[i].ordemServicoID == 0 {
			o.historicoStatus[i].atribuirOrdemServicoID(id)
		}
	}
}

// ValidarTransicaoPara centraliza as invariantes necessárias para uma mudança
// de status. Assim, os próximos fluxos da OS reutilizam as mesmas regras.
func (o *OrdemServico) ValidarTransicaoPara(novo StatusOrdemServico) error {
	if !o.status.PermiteTransicaoPara(novo) {
		return ErrTransicaoStatusInvalida
	}

	if o.status == StatusEmDiagnostico &&
		novo == StatusAguardandoAprovacao &&
		strings.TrimSpace(o.diagnostico) == "" {
		return ErrDiagnosticoObrigatorio
	}

	return nil
}

// IniciarDiagnostico move uma OS recebida para diagnóstico sem preencher o
// diagnóstico e registra quem realizou a transição.
func (o *OrdemServico) IniciarDiagnostico(alteradoPor uint64) error {
	if err := o.ValidarTransicaoPara(StatusEmDiagnostico); err != nil {
		return err
	}

	historico, err := NewHistoricoStatus(StatusEmDiagnostico, alteradoPor, "", time.Now())
	if err != nil {
		return err
	}
	historico.atribuirOrdemServicoID(o.id)

	o.status = StatusEmDiagnostico
	o.diagnostico = ""
	o.dataAtualizacao = historico.AlteradoEm()
	o.historicoStatus = append(o.historicoStatus, historico)
	return nil
}

// InformarDiagnostico registra o diagnóstico enquanto a OS está em análise.
func (o *OrdemServico) InformarDiagnostico(texto string) error {
	if o.status != StatusEmDiagnostico {
		return ErrDiagnosticoStatusInvalido
	}

	diagnostico := strings.TrimSpace(texto)
	if diagnostico == "" {
		return ErrDiagnosticoObrigatorio
	}

	o.diagnostico = diagnostico
	o.dataAtualizacao = time.Now()
	return nil
}

// Entregar move uma OS finalizada para entregue e registra quem realizou a transição.
func (o *OrdemServico) Entregar(alteradoPor uint64) error {
	if err := o.ValidarTransicaoPara(StatusEntregue); err != nil {
		return err
	}

	historico, err := NewHistoricoStatus(StatusEntregue, alteradoPor, "", time.Now())
	if err != nil {
		return err
	}
	historico.atribuirOrdemServicoID(o.id)

	o.status = StatusEntregue
	o.dataAtualizacao = historico.AlteradoEm()
	o.historicoStatus = append(o.historicoStatus, historico)
	return nil
}

func (o *OrdemServico) ID() uint64                 { return o.id }
func (o *OrdemServico) Numero() NumeroOrdemServico { return o.numero }
func (o *OrdemServico) ClienteID() uint64          { return o.clienteID }
func (o *OrdemServico) VeiculoID() uint64          { return o.veiculoID }
func (o *OrdemServico) QuilometragemEntrada() int  { return o.quilometragemEntrada }
func (o *OrdemServico) Status() StatusOrdemServico { return o.status }
func (o *OrdemServico) Diagnostico() string        { return o.diagnostico }
func (o *OrdemServico) Observacoes() string        { return o.observacoes }
func (o *OrdemServico) CriadoPor() uint64          { return o.criadoPor }
func (o *OrdemServico) DataCadastro() time.Time    { return o.dataCadastro }
func (o *OrdemServico) DataAtualizacao() time.Time { return o.dataAtualizacao }

// HistoricoStatus devolve uma cópia para preservar o encapsulamento do agregado.
func (o *OrdemServico) HistoricoStatus() []HistoricoStatus {
	return append([]HistoricoStatus(nil), o.historicoStatus...)
}
