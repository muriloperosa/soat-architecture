package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

var (
	errApprovalTokenConfig  = errors.New("configuração inválida do token de aprovação")
	errApprovalTokenInvalid = errors.New("token de aprovação inválido")
)

const approvalTokenSigningMethod = "HS256"

type approvalTokenClaims struct {
	OrdemServicoID uint64 `json:"ordem_servico_id"`
	ClienteID      uint64 `json:"cliente_id"`
	jwt.RegisteredClaims
}

// ApprovalTokenJWT implementa o token do link de aprovação, isolado do JWT de
// login para permitir segredo e validade próprios.
type ApprovalTokenJWT struct {
	secret []byte
	ttl    time.Duration
}

// NewApprovalTokenJWT cria o provider. O segredo deve ser dedicado a este
// fluxo e o TTL deve ser positivo (24 horas no padrão da aplicação).
func NewApprovalTokenJWT(secret string, ttl time.Duration) (*ApprovalTokenJWT, error) {
	if secret == "" || ttl <= 0 {
		return nil, errApprovalTokenConfig
	}
	return &ApprovalTokenJWT{secret: []byte(secret), ttl: ttl}, nil
}

// GerarToken assina os identificadores da OS e do cliente com expiração.
func (p *ApprovalTokenJWT) GerarToken(ordemServicoID, clienteID uint64) (string, error) {
	if p == nil || len(p.secret) == 0 || p.ttl <= 0 || ordemServicoID == 0 || clienteID == 0 {
		return "", errApprovalTokenConfig
	}

	claims := approvalTokenClaims{
		OrdemServicoID: ordemServicoID,
		ClienteID:      clienteID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(p.ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(p.secret)
}

// ValidarToken verifica assinatura, algoritmo, validade e identificadores.
func (p *ApprovalTokenJWT) ValidarToken(raw string) (shared.ApprovalTokenClaims, error) {
	if p == nil || len(p.secret) == 0 || raw == "" {
		return shared.ApprovalTokenClaims{}, errApprovalTokenInvalid
	}

	claims := &approvalTokenClaims{}
	parser := jwt.NewParser(jwt.WithValidMethods([]string{approvalTokenSigningMethod}), jwt.WithExpirationRequired())
	token, err := parser.ParseWithClaims(raw, claims, func(token *jwt.Token) (any, error) {
		if token.Method.Alg() != approvalTokenSigningMethod {
			return nil, errApprovalTokenInvalid
		}
		return p.secret, nil
	})
	if err != nil || token == nil || !token.Valid || claims.OrdemServicoID == 0 || claims.ClienteID == 0 || claims.ExpiresAt == nil {
		return shared.ApprovalTokenClaims{}, errApprovalTokenInvalid
	}

	return shared.ApprovalTokenClaims{
		OrdemServicoID: claims.OrdemServicoID,
		ClienteID:      claims.ClienteID,
		ExpiresAt:      claims.ExpiresAt.Time,
	}, nil
}
