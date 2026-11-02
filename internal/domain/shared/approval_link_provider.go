package shared

// ApprovalEmailLinks contém as URLs de decisão inseridas no e-mail do orçamento.
type ApprovalEmailLinks struct {
	AprovarURL  string
	RejeitarURL string
}

// ApprovalLinkProvider cria links públicos para um ciclo específico de aprovação.
type ApprovalLinkProvider interface {
	GerarLinksAprovacao(ordemServicoID, clienteID, approvalHistoryID uint64) (ApprovalEmailLinks, error)
}
