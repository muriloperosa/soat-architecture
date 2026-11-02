package email_test

import (
	"errors"
	"testing"

	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/email"
	"github.com/stretchr/testify/require"
)

func TestApprovalEmailLinkProvider_GeraLinksComTokenEURLPublica(t *testing.T) {
	tokens := &approvalLinkTokenStub{token: "assinado+token"}
	provider, err := email.NewApprovalEmailLinkProvider(tokens, "https://oficina.example/")
	require.NoError(t, err)

	links, err := provider.GerarLinksAprovacao(71, 92, 103)
	require.NoError(t, err)
	require.Equal(t, "https://oficina.example/v1/orcamentos/aprovacao?token=assinado%2Btoken&acao=APROVAR", links.AprovarURL)
	require.Equal(t, "https://oficina.example/v1/orcamentos/aprovacao?token=assinado%2Btoken&acao=REJEITAR", links.RejeitarURL)
	require.Equal(t, uint64(71), tokens.ordemServicoID)
	require.Equal(t, uint64(92), tokens.clienteID)
	require.Equal(t, uint64(103), tokens.approvalHistoryID)
}

func TestNewApprovalEmailLinkProvider_RejeitaConfiguracaoInvalida(t *testing.T) {
	tests := []struct {
		name          string
		tokens        shared.ApprovalTokenProvider
		publicBaseURL string
	}{
		{
			name:          "nil token provider",
			publicBaseURL: "https://oficina.example",
		},
		{
			name:          "empty URL",
			tokens:        &approvalLinkTokenStub{},
			publicBaseURL: "",
		},
		{
			name:          "malformed URL",
			tokens:        &approvalLinkTokenStub{},
			publicBaseURL: "http://[::1",
		},
		{
			name:          "unsupported scheme",
			tokens:        &approvalLinkTokenStub{},
			publicBaseURL: "javascript://oficina.example",
		},
		{
			name:          "missing host",
			tokens:        &approvalLinkTokenStub{},
			publicBaseURL: "https:///orcamentos",
		},
		{
			name:          "userinfo",
			tokens:        &approvalLinkTokenStub{},
			publicBaseURL: "https://user:password@oficina.example",
		},
		{
			name:          "query string",
			tokens:        &approvalLinkTokenStub{},
			publicBaseURL: "https://oficina.example?host=invalid",
		},
		{
			name:          "fragment",
			tokens:        &approvalLinkTokenStub{},
			publicBaseURL: "https://oficina.example#section",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := email.NewApprovalEmailLinkProvider(test.tokens, test.publicBaseURL)
			require.Error(t, err)
		})
	}
}

func TestNewApprovalEmailLinkProvider_AcceptaURLComPathBase(t *testing.T) {
	tokens := &approvalLinkTokenStub{token: "assinado+token"}
	provider, err := email.NewApprovalEmailLinkProvider(tokens, "https://oficina.example/app/")
	require.NoError(t, err)

	links, err := provider.GerarLinksAprovacao(71, 92, 103)
	require.NoError(t, err)
	require.Equal(t, "https://oficina.example/app/v1/orcamentos/aprovacao?token=assinado%2Btoken&acao=APROVAR", links.AprovarURL)
}

func TestApprovalEmailLinkProvider_RetornaErroDoTokenProvider(t *testing.T) {
	tokens := &approvalLinkTokenStub{err: errors.New("falha ao assinar")}
	provider, err := email.NewApprovalEmailLinkProvider(tokens, "https://oficina.example")
	require.NoError(t, err)

	_, err = provider.GerarLinksAprovacao(71, 92, 103)
	require.ErrorIs(t, err, tokens.err)
}

type approvalLinkTokenStub struct {
	token             string
	err               error
	ordemServicoID    uint64
	clienteID         uint64
	approvalHistoryID uint64
}

func (s *approvalLinkTokenStub) GerarToken(ordemServicoID, clienteID, approvalHistoryID uint64) (string, error) {
	s.ordemServicoID = ordemServicoID
	s.clienteID = clienteID
	s.approvalHistoryID = approvalHistoryID
	return s.token, s.err
}

func (s *approvalLinkTokenStub) ValidarToken(string) (shared.ApprovalTokenClaims, error) {
	return shared.ApprovalTokenClaims{}, nil
}
