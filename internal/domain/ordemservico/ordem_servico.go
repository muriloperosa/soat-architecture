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

	historicoInicial, err := NewHistoricoStatus(StatusRecebida, criadoPor, "")
	if err != nil {
		return nil, err
	}

	agora := time.Now()

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
