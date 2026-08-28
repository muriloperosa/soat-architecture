package ordemservico_test

import (
	"math"
	"strings"
	"testing"
	"time"

	"github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	"github.com/stretchr/testify/require"
)

func TestNewOrdemServico(t *testing.T) {
	antes := time.Now()
	os, err := ordemservico.NewOrdemServico(
		" OS-2026-0001 ",
		10,
		20,
		52_300,
		" diagnóstico inicial ",
		" cliente aguardando ",
		30,
	)
	depois := time.Now()

	require.NoError(t, err)
	require.Zero(t, os.ID())
	require.Equal(t, "OS-2026-0001", os.Numero().String())
	require.Equal(t, uint64(10), os.ClienteID())
	require.Equal(t, uint64(20), os.VeiculoID())
	require.Equal(t, 52_300, os.QuilometragemEntrada())
	require.Equal(t, ordemservico.StatusRecebida, os.Status())
	require.Equal(t, "diagnóstico inicial", os.Diagnostico())
	require.Equal(t, "cliente aguardando", os.Observacoes())
	require.Equal(t, uint64(30), os.CriadoPor())
	require.False(t, os.DataCadastro().Before(antes))
	require.False(t, os.DataCadastro().After(depois))
	require.Equal(t, os.DataCadastro(), os.DataAtualizacao())

	historico := os.HistoricoStatus()
	require.Len(t, historico, 1)
	require.Equal(t, uint64(0), historico[0].ID())
	require.Equal(t, ordemservico.StatusRecebida, historico[0].Status())
	require.Equal(t, uint64(30), historico[0].AlteradoPor())
	require.Equal(t, os.DataCadastro(), historico[0].AlteradoEm())
	require.Empty(t, historico[0].Motivo())
}

func TestNewOrdemServicoValidaInvariantes(t *testing.T) {
	testes := []struct {
		nome          string
		numero        string
		clienteID     uint64
		veiculoID     uint64
		quilometragem int
		criadoPor     uint64
		erroEsperado  error
	}{
		{"cliente obrigatório", "OS-1", 0, 2, 0, 3, ordemservico.ErrClienteObrigatorio},
		{"veículo obrigatório", "OS-1", 1, 0, 0, 3, ordemservico.ErrVeiculoObrigatorio},
		{"número obrigatório", "   ", 1, 2, 0, 3, ordemservico.ErrNumeroObrigatorio},
		{"número acima do limite", strings.Repeat("A", 51), 1, 2, 0, 3, ordemservico.ErrNumeroInvalido},
		{"quilometragem negativa", "OS-1", 1, 2, -1, 3, ordemservico.ErrQuilometragemEntradaInvalida},
		{"quilometragem acima do INT UNSIGNED", "OS-1", 1, 2, int(uint64(math.MaxUint32) + 1), 3, ordemservico.ErrQuilometragemEntradaInvalida},
		{"responsável obrigatório", "OS-1", 1, 2, 0, 0, ordemservico.ErrCriadoPorObrigatorio},
	}

	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			os, err := ordemservico.NewOrdemServico(
				teste.numero,
				teste.clienteID,
				teste.veiculoID,
				teste.quilometragem,
				"",
				"",
				teste.criadoPor,
			)

			require.Nil(t, os)
			require.ErrorIs(t, err, teste.erroEsperado)
		})
	}
}

func TestStatusOrdemServicoSuportaTodosOsValoresDoFluxo(t *testing.T) {
	statusValidos := []ordemservico.StatusOrdemServico{
		ordemservico.StatusRecebida,
		ordemservico.StatusEmDiagnostico,
		ordemservico.StatusAguardandoAprovacao,
		ordemservico.StatusAprovada,
		ordemservico.StatusRejeitada,
		ordemservico.StatusEmExecucao,
		ordemservico.StatusFinalizada,
		ordemservico.StatusEntregue,
	}

	for _, esperado := range statusValidos {
		status, err := ordemservico.NewStatusOrdemServico(esperado.String())
		require.NoError(t, err)
		require.Equal(t, esperado, status)
		require.True(t, status.IsValid())
	}

	status, err := ordemservico.NewStatusOrdemServico("CANCELADA")
	require.ErrorIs(t, err, ordemservico.ErrStatusInvalido)
	require.Zero(t, status)
}

func TestReidratarOrdemServico(t *testing.T) {
	numero, err := ordemservico.NewNumeroOrdemServico("OS-2026-0042")
	require.NoError(t, err)

	cadastro := time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC)
	atualizacao := cadastro.Add(2 * time.Hour)
	historico := []ordemservico.HistoricoStatus{
		ordemservico.ReidratarHistoricoStatus(7, 42, ordemservico.StatusEmDiagnostico, atualizacao, 3, "início do diagnóstico"),
	}

	os := ordemservico.ReidratarOrdemServico(
		42,
		numero,
		10,
		20,
		52_300,
		ordemservico.StatusEmDiagnostico,
		"ruído no motor",
		"",
		3,
		historico,
		cadastro,
		atualizacao,
	)

	require.Equal(t, uint64(42), os.ID())
	require.Equal(t, numero, os.Numero())
	require.Equal(t, uint64(10), os.ClienteID())
	require.Equal(t, uint64(20), os.VeiculoID())
	require.Equal(t, 52_300, os.QuilometragemEntrada())
	require.Equal(t, ordemservico.StatusEmDiagnostico, os.Status())
	require.Equal(t, "ruído no motor", os.Diagnostico())
	require.Equal(t, uint64(3), os.CriadoPor())
	require.Equal(t, cadastro, os.DataCadastro())
	require.Equal(t, atualizacao, os.DataAtualizacao())
	require.Equal(t, historico, os.HistoricoStatus())

	historico[0] = ordemservico.HistoricoStatus{}
	require.Equal(t, ordemservico.StatusEmDiagnostico, os.HistoricoStatus()[0].Status())
}

func TestAtribuirIDPropagaIdentidadeAoHistoricoInicial(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 0, "", "", 3)
	require.NoError(t, err)

	os.AtribuirID(99)

	require.Equal(t, uint64(99), os.ID())
	require.Equal(t, uint64(99), os.HistoricoStatus()[0].OrdemServicoID())
}

func TestNewHistoricoStatusValidaDados(t *testing.T) {
	_, err := ordemservico.NewHistoricoStatus("INVALIDO", 1, "", time.Now())
	require.ErrorIs(t, err, ordemservico.ErrStatusInvalido)

	_, err = ordemservico.NewHistoricoStatus(ordemservico.StatusRecebida, 0, "", time.Now())
	require.ErrorIs(t, err, ordemservico.ErrResponsavelHistoricoObrigatorio)
}
