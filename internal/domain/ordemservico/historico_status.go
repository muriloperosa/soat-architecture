package ordemservico

import (
	"strings"
	"time"
)

// HistoricoStatus registra um estado assumido pela OS e quem realizou a mudança.
type HistoricoStatus struct {
	id             uint64
	ordemServicoID uint64
	status         StatusOrdemServico
	alteradoEm     time.Time
	alteradoPor    uint64
	motivo         string
}

func NewHistoricoStatus(
	status StatusOrdemServico,
	alteradoPor uint64,
	motivo string,
	alteradoEm time.Time,
) (HistoricoStatus, error) {
	if !status.IsValid() {
		return HistoricoStatus{}, ErrStatusInvalido
	}
	if alteradoPor == 0 {
		return HistoricoStatus{}, ErrResponsavelHistoricoObrigatorio
	}

	return HistoricoStatus{
		status:      status,
		alteradoEm:  alteradoEm,
		alteradoPor: alteradoPor,
		motivo:      strings.TrimSpace(motivo),
	}, nil
}

func ReidratarHistoricoStatus(
	id, ordemServicoID uint64,
	status StatusOrdemServico,
	alteradoEm time.Time,
	alteradoPor uint64,
	motivo string,
) HistoricoStatus {
	return HistoricoStatus{
		id:             id,
		ordemServicoID: ordemServicoID,
		status:         status,
		alteradoEm:     alteradoEm,
		alteradoPor:    alteradoPor,
		motivo:         motivo,
	}
}

func (h HistoricoStatus) ID() uint64                 { return h.id }
func (h HistoricoStatus) OrdemServicoID() uint64     { return h.ordemServicoID }
func (h HistoricoStatus) Status() StatusOrdemServico { return h.status }
func (h HistoricoStatus) AlteradoEm() time.Time      { return h.alteradoEm }
func (h HistoricoStatus) AlteradoPor() uint64        { return h.alteradoPor }
func (h HistoricoStatus) Motivo() string             { return h.motivo }

func (h *HistoricoStatus) atribuirOrdemServicoID(id uint64) {
	h.ordemServicoID = id
}
