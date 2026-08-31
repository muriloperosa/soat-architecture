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

// PermiteTransicaoPara informa se o status pode avançar para o próximo
// estado segundo o fluxo de negócio da Ordem de Serviço.
func (s StatusOrdemServico) PermiteTransicaoPara(novo StatusOrdemServico) bool {
	switch s {
	case StatusRecebida:
		return novo == StatusEmDiagnostico
	case StatusEmDiagnostico:
		return novo == StatusAguardandoAprovacao
	case StatusAguardandoAprovacao:
		return novo == StatusAprovada || novo == StatusRejeitada
	case StatusRejeitada:
		return novo == StatusAguardandoAprovacao
	case StatusAprovada:
		return novo == StatusEmExecucao || novo == StatusAguardandoAprovacao
	case StatusEmExecucao:
		return novo == StatusFinalizada
	case StatusFinalizada:
		return novo == StatusEntregue
	default:
		return false
	}
}

var todosStatus = []StatusOrdemServico{
	StatusRecebida,
	StatusEmDiagnostico,
	StatusAguardandoAprovacao,
	StatusAprovada,
	StatusRejeitada,
	StatusEmExecucao,
	StatusFinalizada,
	StatusEntregue,
}

// ExisteCaminhoValido informa se existe uma sequência de transições diretas
// válidas (PermiteTransicaoPara) ligando s a destino. Usa busca em largura
// sobre o grafo de transições, suportando o ciclo AguardandoAprovacao<->Rejeitada
// sem entrar em loop infinito.
func (s StatusOrdemServico) ExisteCaminhoValido(destino StatusOrdemServico) bool {
	if s == destino {
		return false
	}

	visitados := map[StatusOrdemServico]bool{s: true}
	fila := []StatusOrdemServico{s}

	for len(fila) > 0 {
		atual := fila[0]
		fila = fila[1:]

		for _, proximo := range todosStatus {
			if !atual.PermiteTransicaoPara(proximo) {
				continue
			}
			if proximo == destino {
				return true
			}
			if !visitados[proximo] {
				visitados[proximo] = true
				fila = append(fila, proximo)
			}
		}
	}

	return false
}

func (s StatusOrdemServico) String() string { return string(s) }
