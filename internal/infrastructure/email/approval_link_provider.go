package email

import (
	"errors"
	"net/url"
	"strings"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

var errApprovalLinkConfig = errors.New("configuração inválida dos links de aprovação")

type ApprovalEmailLinkProvider struct {
	tokens        shared.ApprovalTokenProvider
	publicBaseURL string
}

var _ shared.ApprovalLinkProvider = (*ApprovalEmailLinkProvider)(nil)

// NewApprovalEmailLinkProvider monta links com base em uma URL configurada,
// sem depender de headers da requisição que disparou a notificação.
func NewApprovalEmailLinkProvider(
	tokens shared.ApprovalTokenProvider,
	publicBaseURL string,
) (*ApprovalEmailLinkProvider, error) {
	parsedURL, err := url.Parse(strings.TrimRight(publicBaseURL, "/"))
	if tokens == nil || err != nil || !validPublicBaseURL(parsedURL) {
		return nil, errApprovalLinkConfig
	}
	return &ApprovalEmailLinkProvider{
		tokens:        tokens,
		publicBaseURL: strings.TrimRight(publicBaseURL, "/"),
	}, nil
}

func validPublicBaseURL(u *url.URL) bool {
	return u != nil &&
		(u.Scheme == "http" || u.Scheme == "https") &&
		u.Host != "" &&
		u.User == nil &&
		u.RawQuery == "" &&
		u.Fragment == ""
}

func (p *ApprovalEmailLinkProvider) GerarLinksAprovacao(ordemServicoID, clienteID, approvalHistoryID uint64) (shared.ApprovalEmailLinks, error) {
	if p == nil || p.tokens == nil {
		return shared.ApprovalEmailLinks{}, errApprovalLinkConfig
	}
	token, err := p.tokens.GerarToken(ordemServicoID, clienteID, approvalHistoryID)
	if err != nil {
		return shared.ApprovalEmailLinks{}, err
	}
	endpoint := p.publicBaseURL + "/v1/orcamentos/aprovacao?token=" + url.QueryEscape(token)
	return shared.ApprovalEmailLinks{
		AprovarURL:  endpoint + "&acao=APROVAR",
		RejeitarURL: endpoint + "&acao=REJEITAR",
	}, nil
}
