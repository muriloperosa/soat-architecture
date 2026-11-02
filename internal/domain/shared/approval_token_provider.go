package shared

import "time"

// ApprovalTokenClaims identifica a OS, o cliente e o ciclo autorizados pelo
// link de aprovação por e-mail. A expiração é devolvida ao adapter consumidor.
type ApprovalTokenClaims struct {
	OrdemServicoID    uint64
	ClienteID         uint64
	ApprovalHistoryID uint64
	ExpiresAt         time.Time
}

// ApprovalTokenProvider gera e valida tokens temporários do canal de aprovação
// por e-mail. A implementação deve usar um segredo próprio, separado do JWT de
// autenticação da aplicação.
type ApprovalTokenProvider interface {
	GerarToken(ordemServicoID, clienteID, approvalHistoryID uint64) (string, error)
	ValidarToken(token string) (ApprovalTokenClaims, error)
}
