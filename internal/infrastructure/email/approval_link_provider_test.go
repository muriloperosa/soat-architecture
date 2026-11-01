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

	links, err := provider.GerarLinksAprovacao(71, 92)
	require.NoError(t, err)
	require.Equal(t, "https://oficina.example/v1/orcamentos/aprovacao?token=assinado%2Btoken&acao=APROVAR", links.AprovarURL)
	require.Equal(t, "https://oficina.example/v1/orcamentos/aprovacao?token=assinado%2Btoken&acao=REJEITAR", links.RejeitarURL)
	require.Equal(t, uint64(71), tokens.ordemServicoID)
	require.Equal(t, uint64(92), tokens.clienteID)
}

func TestNewApprovalEmailLinkProvider_RejeitaConfiguracaoInvalida(t *testing.T) {
	tokens := &approvalLinkTokenStub{}
	_, err := email.NewApprovalEmailLinkProvider(tokens, "")
	require.Error(t, err)
	_, err = email.NewApprovalEmailLinkProvider(nil, "https://oficina.example")
	require.Error(t, err)
	_, err = email.NewApprovalEmailLinkProvider(tokens, "https://oficina.example?host=invalid")
	require.Error(t, err)
	_, err = email.NewApprovalEmailLinkProvider(tokens, "javascript://oficina.example")
	require.Error(t, err)
}

func TestApprovalEmailLinkProvider_RetornaErroDoTokenProvider(t *testing.T) {
	tokens := &approvalLinkTokenStub{err: errors.New("falha ao assinar")}
	provider, err := email.NewApprovalEmailLinkProvider(tokens, "https://oficina.example")
	require.NoError(t, err)

	_, err = provider.GerarLinksAprovacao(71, 92)
	require.ErrorIs(t, err, tokens.err)
}

type approvalLinkTokenStub struct {
	token          string
	err            error
	ordemServicoID uint64
	clienteID      uint64
}

func (s *approvalLinkTokenStub) GerarToken(ordemServicoID, clienteID uint64) (string, error) {
	s.ordemServicoID = ordemServicoID
	s.clienteID = clienteID
	return s.token, s.err
}

func (s *approvalLinkTokenStub) ValidarToken(string) (shared.ApprovalTokenClaims, error) {
	return shared.ApprovalTokenClaims{}, nil
}
