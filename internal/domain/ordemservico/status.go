package ordemservico

// StatusOrdemServico representa os estados possíveis da Ordem de Serviço.
type StatusOrdemServico string

const (
	StatusRecebida            StatusOrdemServico = "RECEBIDA"
	StatusEmDiagnostico       StatusOrdemServico = "EM_DIAGNOSTICO"
	StatusAguardandoAprovacao StatusOrdemServico = "AGUARDANDO_APROVACAO"
	StatusAprovada            StatusOrdemServico = "APROVADA"
	StatusRejeitada           StatusOrdemServico = "REJEITADA"
	StatusEmExecucao          StatusOrdemServico = "EM_EXECUCAO"
	StatusFinalizada          StatusOrdemServico = "FINALIZADA"
	StatusEntregue            StatusOrdemServico = "ENTREGUE"
)

// NewStatusOrdemServico restaura um status a partir de sua representação textual.
func NewStatusOrdemServico(valor string) (StatusOrdemServico, error) {
	status := StatusOrdemServico(valor)
	if !status.IsValid() {
		return "", ErrStatusInvalido
	}

	return status, nil
}

func (s StatusOrdemServico) IsValid() bool {
	switch s {
	case StatusRecebida,
		StatusEmDiagnostico,
		StatusAguardandoAprovacao,
		StatusAprovada,
		StatusRejeitada,
		StatusEmExecucao,
		StatusFinalizada,
		StatusEntregue:
		return true
	default:
		return false
	}
}

func (s StatusOrdemServico) String() string { return string(s) }
