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

func TestIniciarDiagnostico_ComSucesso(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "diagnóstico anterior", "", 3)
	require.NoError(t, err)
	os.AtribuirID(42)
	antes := os.DataAtualizacao()

	err = os.IniciarDiagnostico(7)

	require.NoError(t, err)
	require.Equal(t, ordemservico.StatusEmDiagnostico, os.Status())
	require.Empty(t, os.Diagnostico())
	require.False(t, os.DataAtualizacao().Before(antes))
	historico := os.HistoricoStatus()
	require.Len(t, historico, 2)
	require.Equal(t, ordemservico.StatusEmDiagnostico, historico[1].Status())
	require.Equal(t, uint64(42), historico[1].OrdemServicoID())
	require.Equal(t, uint64(7), historico[1].AlteradoPor())
	require.Empty(t, historico[1].Motivo())
	require.Equal(t, os.DataAtualizacao(), historico[1].AlteradoEm())
}

func TestIniciarDiagnostico_SomenteOSRecebida(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)
	require.NoError(t, os.IniciarDiagnostico(7))

	err = os.IniciarDiagnostico(7)

	require.ErrorIs(t, err, ordemservico.ErrTransicaoStatusInvalida)
	require.Equal(t, ordemservico.StatusEmDiagnostico, os.Status())
	require.Len(t, os.HistoricoStatus(), 2)
}

func TestIniciarDiagnostico_ResponsavelObrigatorio(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)

	err = os.IniciarDiagnostico(0)

	require.ErrorIs(t, err, ordemservico.ErrResponsavelHistoricoObrigatorio)
	require.Equal(t, ordemservico.StatusRecebida, os.Status())
	require.Len(t, os.HistoricoStatus(), 1)
}

func TestInformarDiagnostico_ComSucesso(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)
	require.NoError(t, os.IniciarDiagnostico(7))
	quantidadeHistoricos := len(os.HistoricoStatus())

	err = os.InformarDiagnostico("  Falha na bomba de combustível  ")

	require.NoError(t, err)
	require.Equal(t, "Falha na bomba de combustível", os.Diagnostico())
	require.Len(t, os.HistoricoStatus(), quantidadeHistoricos)
}

func TestInformarDiagnostico_ValidaEstadoEConteudo(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)

	err = os.InformarDiagnostico("Falha na bomba")
	require.ErrorIs(t, err, ordemservico.ErrDiagnosticoStatusInvalido)

	require.NoError(t, os.IniciarDiagnostico(7))
	err = os.InformarDiagnostico("   ")
	require.ErrorIs(t, err, ordemservico.ErrDiagnosticoObrigatorio)
	require.Empty(t, os.Diagnostico())
}

func TestValidarTransicaoParaAguardandoAprovacao_ExigeDiagnostico(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)
	require.NoError(t, os.IniciarDiagnostico(7))

	err = os.ValidarTransicaoPara(ordemservico.StatusAguardandoAprovacao)
	require.ErrorIs(t, err, ordemservico.ErrDiagnosticoObrigatorio)

	require.NoError(t, os.InformarDiagnostico("Falha na bomba de combustível"))
	require.NoError(t, os.ValidarTransicaoPara(ordemservico.StatusAguardandoAprovacao))
}

func TestEnviarParaAprovacao_ComSucesso(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)
	os.AtribuirID(42)
	require.NoError(t, os.IniciarDiagnostico(7))
	require.NoError(t, os.InformarDiagnostico("Falha na bomba de combustível"))
	antes := os.DataAtualizacao()

	err = os.EnviarParaAprovacao(7)

	require.NoError(t, err)
	require.Equal(t, ordemservico.StatusAguardandoAprovacao, os.Status())
	require.False(t, os.DataAtualizacao().Before(antes))
	historico := os.HistoricoStatus()
	require.Len(t, historico, 3)
	require.Equal(t, ordemservico.StatusAguardandoAprovacao, historico[2].Status())
	require.Equal(t, uint64(42), historico[2].OrdemServicoID())
	require.Equal(t, uint64(7), historico[2].AlteradoPor())
}

func TestEnviarParaAprovacao_ExigeDiagnostico(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)
	require.NoError(t, os.IniciarDiagnostico(7))

	err = os.EnviarParaAprovacao(7)

	require.ErrorIs(t, err, ordemservico.ErrDiagnosticoObrigatorio)
	require.Equal(t, ordemservico.StatusEmDiagnostico, os.Status())
}

func TestEnviarParaAprovacao_SomenteOSEmDiagnostico(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)

	err = os.EnviarParaAprovacao(7)

	require.ErrorIs(t, err, ordemservico.ErrTransicaoStatusInvalida)
	require.Equal(t, ordemservico.StatusRecebida, os.Status())
}

func osAprovada(t *testing.T, id uint64) *ordemservico.OrdemServico {
	t.Helper()
	numero, err := ordemservico.NewNumeroOrdemServico("OS-2026-0077")
	require.NoError(t, err)

	cadastro := time.Date(2026, time.August, 10, 9, 0, 0, 0, time.UTC)
	atualizacao := cadastro.Add(4 * time.Hour)
	historico := []ordemservico.HistoricoStatus{
		ordemservico.ReidratarHistoricoStatus(1, id, ordemservico.StatusRecebida, cadastro, 3, ""),
		ordemservico.ReidratarHistoricoStatus(2, id, ordemservico.StatusAprovada, atualizacao, 3, ""),
	}

	return ordemservico.ReidratarOrdemServico(
		id,
		numero,
		10,
		20,
		52_300,
		ordemservico.StatusAprovada,
		"motor revisado",
		"",
		3,
		historico,
		cadastro,
		atualizacao,
	)
}

func TestIniciarExecucao_ComSucesso(t *testing.T) {
	os := osAprovada(t, 42)
	antes := os.DataAtualizacao()

	err := os.IniciarExecucao(7)

	require.NoError(t, err)
	require.Equal(t, ordemservico.StatusEmExecucao, os.Status())
	require.False(t, os.DataAtualizacao().Before(antes))
	historico := os.HistoricoStatus()
	require.Len(t, historico, 3)
	require.Equal(t, ordemservico.StatusEmExecucao, historico[2].Status())
	require.Equal(t, uint64(42), historico[2].OrdemServicoID())
	require.Equal(t, uint64(7), historico[2].AlteradoPor())
	require.Empty(t, historico[2].Motivo())
	require.Equal(t, os.DataAtualizacao(), historico[2].AlteradoEm())
}

func TestIniciarExecucao_SomenteOSAprovada(t *testing.T) {
	testes := []struct {
		nome   string
		status ordemservico.StatusOrdemServico
	}{
		{"recebida", ordemservico.StatusRecebida},
		{"rejeitada", ordemservico.StatusRejeitada},
		{"aguardando aprovação", ordemservico.StatusAguardandoAprovacao},
	}

	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			os := osAprovada(t, 42)
			os = ordemservico.ReidratarOrdemServico(
				os.ID(), os.Numero(), os.ClienteID(), os.VeiculoID(), os.QuilometragemEntrada(),
				teste.status, os.Diagnostico(), os.Observacoes(), os.CriadoPor(),
				os.HistoricoStatus(), os.DataCadastro(), os.DataAtualizacao(),
			)

			err := os.IniciarExecucao(7)

			require.ErrorIs(t, err, ordemservico.ErrTransicaoStatusInvalida)
			require.Equal(t, teste.status, os.Status())
		})
	}
}

func TestIniciarExecucao_ResponsavelObrigatorio(t *testing.T) {
	os := osAprovada(t, 42)

	err := os.IniciarExecucao(0)

	require.ErrorIs(t, err, ordemservico.ErrResponsavelHistoricoObrigatorio)
	require.Equal(t, ordemservico.StatusAprovada, os.Status())
	require.Len(t, os.HistoricoStatus(), 2)
}

func osEmExecucao(t *testing.T, id uint64) *ordemservico.OrdemServico {
	t.Helper()
	os := osAprovada(t, id)
	require.NoError(t, os.IniciarExecucao(7))
	return os
}

func TestFinalizar_ComSucesso(t *testing.T) {
	os := osEmExecucao(t, 42)
	antes := os.DataAtualizacao()

	err := os.Finalizar(9)

	require.NoError(t, err)
	require.Equal(t, ordemservico.StatusFinalizada, os.Status())
	require.False(t, os.DataAtualizacao().Before(antes))
	historico := os.HistoricoStatus()
	require.Len(t, historico, 4)
	require.Equal(t, ordemservico.StatusFinalizada, historico[3].Status())
	require.Equal(t, uint64(42), historico[3].OrdemServicoID())
	require.Equal(t, uint64(9), historico[3].AlteradoPor())
	require.Empty(t, historico[3].Motivo())
	require.Equal(t, os.DataAtualizacao(), historico[3].AlteradoEm())
}

func TestFinalizar_SomenteOSEmExecucao(t *testing.T) {
	testes := []struct {
		nome   string
		status ordemservico.StatusOrdemServico
	}{
		{"recebida", ordemservico.StatusRecebida},
		{"aprovada", ordemservico.StatusAprovada},
		{"finalizada", ordemservico.StatusFinalizada},
	}

	for _, teste := range testes {
		t.Run(teste.nome, func(t *testing.T) {
			os := osAprovada(t, 42)
			os = ordemservico.ReidratarOrdemServico(
				os.ID(), os.Numero(), os.ClienteID(), os.VeiculoID(), os.QuilometragemEntrada(),
				teste.status, os.Diagnostico(), os.Observacoes(), os.CriadoPor(),
				os.HistoricoStatus(), os.DataCadastro(), os.DataAtualizacao(),
			)

			err := os.Finalizar(7)

			require.ErrorIs(t, err, ordemservico.ErrTransicaoStatusInvalida)
			require.Equal(t, teste.status, os.Status())
		})
	}
}

func TestFinalizar_ResponsavelObrigatorio(t *testing.T) {
	os := osEmExecucao(t, 42)

	err := os.Finalizar(0)

	require.ErrorIs(t, err, ordemservico.ErrResponsavelHistoricoObrigatorio)
	require.Equal(t, ordemservico.StatusEmExecucao, os.Status())
	require.Len(t, os.HistoricoStatus(), 3)
}

func osFinalizada(t *testing.T, id uint64) *ordemservico.OrdemServico {
	t.Helper()
	numero, err := ordemservico.NewNumeroOrdemServico("OS-2026-0099")
	require.NoError(t, err)

	cadastro := time.Date(2026, time.August, 1, 9, 0, 0, 0, time.UTC)
	atualizacao := cadastro.Add(3 * time.Hour)
	historico := []ordemservico.HistoricoStatus{
		ordemservico.ReidratarHistoricoStatus(1, id, ordemservico.StatusRecebida, cadastro, 3, ""),
		ordemservico.ReidratarHistoricoStatus(2, id, ordemservico.StatusFinalizada, atualizacao, 3, ""),
	}

	return ordemservico.ReidratarOrdemServico(
		id,
		numero,
		10,
		20,
		52_300,
		ordemservico.StatusFinalizada,
		"motor revisado",
		"",
		3,
		historico,
		cadastro,
		atualizacao,
	)
}

func TestEntregar_ComSucesso(t *testing.T) {
	os := osFinalizada(t, 42)
	antes := os.DataAtualizacao()

	err := os.Entregar(7)

	require.NoError(t, err)
	require.Equal(t, ordemservico.StatusEntregue, os.Status())
	require.False(t, os.DataAtualizacao().Before(antes))
	historico := os.HistoricoStatus()
	require.Len(t, historico, 3)
	require.Equal(t, ordemservico.StatusEntregue, historico[2].Status())
	require.Equal(t, uint64(42), historico[2].OrdemServicoID())
	require.Equal(t, uint64(7), historico[2].AlteradoPor())
	require.Empty(t, historico[2].Motivo())
	require.Equal(t, os.DataAtualizacao(), historico[2].AlteradoEm())
}

func TestEntregar_SomenteOSFinalizada(t *testing.T) {
	os, err := ordemservico.NewOrdemServico("OS-1", 1, 2, 100, "", "", 3)
	require.NoError(t, err)

	err = os.Entregar(7)

	require.ErrorIs(t, err, ordemservico.ErrTransicaoStatusInvalida)
	require.Equal(t, ordemservico.StatusRecebida, os.Status())
	require.Len(t, os.HistoricoStatus(), 1)
}

func TestEntregar_ResponsavelObrigatorio(t *testing.T) {
	os := osFinalizada(t, 42)

	err := os.Entregar(0)

	require.ErrorIs(t, err, ordemservico.ErrResponsavelHistoricoObrigatorio)
	require.Equal(t, ordemservico.StatusFinalizada, os.Status())
	require.Len(t, os.HistoricoStatus(), 2)
}

func TestEntregar_NaoPermiteTransicaoDeEntregueParaExecucao(t *testing.T) {
	os := osFinalizada(t, 42)
	require.NoError(t, os.Entregar(7))

	err := os.Entregar(7)

	require.ErrorIs(t, err, ordemservico.ErrTransicaoStatusInvalida)
	require.Equal(t, ordemservico.StatusEntregue, os.Status())
	require.False(t, ordemservico.StatusEntregue.PermiteTransicaoPara(ordemservico.StatusEmExecucao))
}

func TestStatusOrdemServicoDefineTransicoesPermitidas(t *testing.T) {
	permitidas := map[ordemservico.StatusOrdemServico][]ordemservico.StatusOrdemServico{
		ordemservico.StatusRecebida:            {ordemservico.StatusEmDiagnostico},
		ordemservico.StatusEmDiagnostico:       {ordemservico.StatusAguardandoAprovacao},
		ordemservico.StatusAguardandoAprovacao: {ordemservico.StatusAprovada, ordemservico.StatusRejeitada},
		ordemservico.StatusRejeitada:           {ordemservico.StatusAguardandoAprovacao},
		ordemservico.StatusAprovada:            {ordemservico.StatusEmExecucao},
		ordemservico.StatusEmExecucao:          {ordemservico.StatusFinalizada},
		ordemservico.StatusFinalizada:          {ordemservico.StatusEntregue},
	}

	for atual, proximos := range permitidas {
		for _, proximo := range proximos {
			require.True(t, atual.PermiteTransicaoPara(proximo), "%s deveria permitir %s", atual, proximo)
		}
	}

	require.False(t, ordemservico.StatusRecebida.PermiteTransicaoPara(ordemservico.StatusFinalizada))
	require.False(t, ordemservico.StatusEntregue.PermiteTransicaoPara(ordemservico.StatusRecebida))
}
