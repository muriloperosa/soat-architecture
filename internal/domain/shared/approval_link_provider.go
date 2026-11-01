package shared

// ApprovalEmailLinks contém as URLs de decisão inseridas no e-mail do orçamento.
type ApprovalEmailLinks struct {
	AprovarURL  string
	RejeitarURL string
}

// ApprovalLinkProvider cria links públicos de decisão para uma OS e seu cliente.
type ApprovalLinkProvider interface {
	GerarLinksAprovacao(ordemServicoID, clienteID uint64) (ApprovalEmailLinks, error)
}
