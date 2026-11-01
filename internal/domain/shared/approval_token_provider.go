package shared

import "time"

// ApprovalTokenClaims identifica a OS e o cliente autorizados pelo link de
// aprovação por e-mail. A expiração é devolvida para o adapter consumidor.
type ApprovalTokenClaims struct {
	OrdemServicoID uint64
	ClienteID      uint64
	ExpiresAt      time.Time
}

// ApprovalTokenProvider gera e valida tokens temporários do canal de aprovação
// por e-mail. A implementação deve usar um segredo próprio, separado do JWT de
// autenticação da aplicação.
type ApprovalTokenProvider interface {
	GerarToken(ordemServicoID, clienteID uint64) (string, error)
	ValidarToken(token string) (ApprovalTokenClaims, error)
}
