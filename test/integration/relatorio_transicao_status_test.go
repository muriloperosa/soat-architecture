//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	domainordemservico "github.com/muriloperosa/soat-architecture/internal/domain/ordemservico"
	domainrelatorio "github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httprelatorio "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/relatorio"
	persistrelatorio "github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql/relatorio"
)

// seedOrdemServicoComNumero insere uma OS diretamente via SQL (sem passar pelos use
// cases, que hoje só cobrem até informar-diagnóstico), pra permitir simular
// livremente sequências de historicos_status com timestamps arbitrários.
func seedOrdemServicoComNumero(t *testing.T, numero string, clienteID, veiculoID, criadoPor uint64) uint64 {
	t.Helper()
	if err := testDB.Exec(
		`INSERT INTO ordens_servico (numero, cliente_id, veiculo_id, quilometragem_entrada, status, criado_por)
		 VALUES (?, ?, ?, ?, 'RECEBIDA', ?)`,
		numero, clienteID, veiculoID, 50_000, criadoPor,
	).Error; err != nil {
		t.Fatalf("erro ao semear ordem de serviço: %v", err)
	}

	var id uint64
	if err := testDB.Raw(`SELECT id FROM ordens_servico WHERE numero = ?`, numero).Scan(&id).Error; err != nil {
		t.Fatalf("erro ao buscar id da ordem de serviço semeada: %v", err)
	}
	return id
}

func seedHistorico(t *testing.T, osID uint64, status string, alteradoPor uint64, alteradoEm time.Time) {
	t.Helper()
	if err := testDB.Exec(
		`INSERT INTO historicos_status (ordem_servico_id, status, alterado_por, alterado_em) VALUES (?, ?, ?, ?)`,
		osID, status, alteradoPor, alteradoEm,
	).Error; err != nil {
		t.Fatalf("erro ao semear histórico de status: %v", err)
	}
}

func seedClienteEVeiculo(t *testing.T, admin uint64, sufixo string) (clienteID, veiculoID uint64) {
	t.Helper()
	cliente, err := testContainer.CriarClienteUseCase.Executar(context.Background(), appcliente.CriarClienteInput{
		Documento: "52998224725",
		Tipo:      domaincliente.TipoPessoaFisica,
		Nome:      "Cliente " + sufixo,
		Email:     "cliente" + sufixo + "@email.com",
		Telefone:  "11999998888",
		Senha:     "senhaCliente123",
		CriadoPor: admin,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}

	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(context.Background(), appveiculo.CadastrarVeiculoInput{
		Placa:              "ABC1" + sufixo,
		Marca:              "Fiat",
		Modelo:             "Uno",
		QuilometragemAtual: 50_000,
		Ano:                2020,
		Cor:                "Prata",
		CriadoPor:          admin,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	return cliente.ID, veiculo.ID
}

func TestRelatorioRepository_CalcularTransicaoStatus(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	clienteID, veiculoID := seedClienteEVeiculo(t, admin.ID, "D01")

	repo := persistrelatorio.NewRelatorioRepository(testDB)
	base := time.Date(2026, time.March, 1, 8, 0, 0, 0, time.UTC)

	// OS 1: completou RECEBIDA -> ENTREGUE dentro do período, em 1h (salto de várias etapas).
	os1 := seedOrdemServicoComNumero(t, "OS-REL-0001", clienteID, veiculoID, admin.ID)
	seedHistorico(t, os1, "RECEBIDA", admin.ID, base)
	seedHistorico(t, os1, "EM_DIAGNOSTICO", admin.ID, base.Add(10*time.Minute))
	seedHistorico(t, os1, "AGUARDANDO_APROVACAO", admin.ID, base.Add(20*time.Minute))
	seedHistorico(t, os1, "APROVADA", admin.ID, base.Add(30*time.Minute))
	seedHistorico(t, os1, "EM_EXECUCAO", admin.ID, base.Add(40*time.Minute))
	seedHistorico(t, os1, "FINALIZADA", admin.ID, base.Add(50*time.Minute))
	seedHistorico(t, os1, "ENTREGUE", admin.ID, base.Add(60*time.Minute))

	// OS 2: completou RECEBIDA -> ENTREGUE dentro do período, em 3h.
	os2 := seedOrdemServicoComNumero(t, "OS-REL-0002", clienteID, veiculoID, admin.ID)
	seedHistorico(t, os2, "RECEBIDA", admin.ID, base.Add(2*time.Hour))
	seedHistorico(t, os2, "ENTREGUE", admin.ID, base.Add(5*time.Hour))

	// OS 3: chegou em RECEBIDA mas nunca chegou em ENTREGUE — não deve contar.
	os3 := seedOrdemServicoComNumero(t, "OS-REL-0003", clienteID, veiculoID, admin.ID)
	seedHistorico(t, os3, "RECEBIDA", admin.ID, base)
	seedHistorico(t, os3, "EM_DIAGNOSTICO", admin.ID, base.Add(10*time.Minute))

	// OS 4: chegou em ENTREGUE antes de RECEBIDA (ordem invertida no histórico,
	// dado inconsistente hipotético) — self-join exige alterado_em > t_from, não deve contar.
	os4 := seedOrdemServicoComNumero(t, "OS-REL-0004", clienteID, veiculoID, admin.ID)
	seedHistorico(t, os4, "ENTREGUE", admin.ID, base.Add(-time.Hour))
	seedHistorico(t, os4, "RECEBIDA", admin.ID, base)

	// OS 5: completou a transição, mas fora do período de busca (t_to fora do range).
	os5 := seedOrdemServicoComNumero(t, "OS-REL-0005", clienteID, veiculoID, admin.ID)
	seedHistorico(t, os5, "RECEBIDA", admin.ID, base.Add(48*time.Hour))
	seedHistorico(t, os5, "ENTREGUE", admin.ID, base.Add(49*time.Hour))

	periodo, err := domainrelatorio.NewPeriodo(base.Add(-24*time.Hour), base.Add(24*time.Hour))
	if err != nil {
		t.Fatalf("erro ao montar período: %v", err)
	}

	resultado, err := repo.CalcularTransicaoStatus(context.Background(), domainrelatorio.CalcularTransicaoStatusParams{
		FromStatus: domainordemservico.StatusRecebida,
		ToStatus:   domainordemservico.StatusEntregue,
		Periodo:    periodo,
	})
	if err != nil {
		t.Fatalf("erro ao calcular transição de status: %v", err)
	}

	if resultado.TotalOrdens != 2 {
		t.Fatalf("esperado total 2 (os1 e os2), encontrado %d", resultado.TotalOrdens)
	}
	if resultado.DuracaoMinima != time.Hour {
		t.Fatalf("esperada duração mínima 1h, encontrada %v", resultado.DuracaoMinima)
	}
	if resultado.DuracaoMaxima != 3*time.Hour {
		t.Fatalf("esperada duração máxima 3h, encontrada %v", resultado.DuracaoMaxima)
	}
	if resultado.DuracaoMedia != 2*time.Hour {
		t.Fatalf("esperada duração média 2h, encontrada %v", resultado.DuracaoMedia)
	}
}

func TestRelatorioRepository_CalcularTransicaoStatus_SemOrdensNoPeriodo(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	clienteID, veiculoID := seedClienteEVeiculo(t, admin.ID, "D02")

	repo := persistrelatorio.NewRelatorioRepository(testDB)
	base := time.Date(2026, time.April, 1, 8, 0, 0, 0, time.UTC)

	os1 := seedOrdemServicoComNumero(t, "OS-REL-0006", clienteID, veiculoID, admin.ID)
	seedHistorico(t, os1, "RECEBIDA", admin.ID, base)
	seedHistorico(t, os1, "ENTREGUE", admin.ID, base.Add(time.Hour))

	periodo, err := domainrelatorio.NewPeriodo(base.Add(-72*time.Hour), base.Add(-48*time.Hour))
	if err != nil {
		t.Fatalf("erro ao montar período: %v", err)
	}

	resultado, err := repo.CalcularTransicaoStatus(context.Background(), domainrelatorio.CalcularTransicaoStatusParams{
		FromStatus: domainordemservico.StatusRecebida,
		ToStatus:   domainordemservico.StatusEntregue,
		Periodo:    periodo,
	})
	if err != nil {
		t.Fatalf("erro ao calcular transição de status: %v", err)
	}

	if resultado.TotalOrdens != 0 {
		t.Fatalf("esperado total 0, encontrado %d", resultado.TotalOrdens)
	}
	if resultado.DuracaoMedia != 0 || resultado.DuracaoMinima != 0 || resultado.DuracaoMaxima != 0 {
		t.Fatalf("esperadas durações zeradas, encontrado media=%v min=%v max=%v",
			resultado.DuracaoMedia, resultado.DuracaoMinima, resultado.DuracaoMaxima)
	}
}

func TestRelatorioTransicaoStatusRoute_ComSucesso(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	clienteID, veiculoID := seedClienteEVeiculo(t, admin.ID, "E01")

	agora := time.Now().UTC()
	os1 := seedOrdemServicoComNumero(t, "OS-REL-HTTP-0001", clienteID, veiculoID, admin.ID)
	seedHistorico(t, os1, "RECEBIDA", admin.ID, agora.Add(-90*time.Minute))
	seedHistorico(t, os1, "ENTREGUE", admin.ID, agora.Add(-30*time.Minute))

	ontem := agora.AddDate(0, 0, -1).Format("2006-01-02")
	hoje := agora.Format("2006-01-02")
	var response httprelatorio.TransicaoStatusResponse
	rec := doRequest(t, http.MethodGet,
		"/v1/relatorios/ordens-servico/transicao-status?start_date="+ontem+"&final_date="+hoje+"&from_status=RECEBIDA&to_status=ENTREGUE&unit=m",
		loginAdmin.AccessToken, nil, &response,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("relatório: status %d, body %q", rec.Code, rec.Body.String())
	}
	if response.TotalOrdensServico != 1 {
		t.Fatalf("esperado total 1, encontrado %d", response.TotalOrdensServico)
	}
	if response.Unidade != "m" || response.TempoMedio != 60 {
		t.Fatalf("resposta inesperada: %+v", response)
	}
}

func TestRelatorioTransicaoStatusRoute_SemToken_Retorna401(t *testing.T) {
	resetDB(t)
	hoje := time.Now().UTC().Format("2006-01-02")

	rec := doRequest(t, http.MethodGet,
		"/v1/relatorios/ordens-servico/transicao-status?start_date="+hoje+"&final_date="+hoje+"&from_status=RECEBIDA&to_status=ENTREGUE",
		"", nil, nil,
	)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("esperado 401 sem token, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestRelatorioTransicaoStatusRoute_PapelNaoAdmin_Retorna403(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Mecânico Oficina", "mecanico@oficina.com", "senha123", shared.PapelMecanico)
	loginMecanico := doLogin(t, "mecanico@oficina.com", "senha123")
	hoje := time.Now().UTC().Format("2006-01-02")

	rec := doRequest(t, http.MethodGet,
		"/v1/relatorios/ordens-servico/transicao-status?start_date="+hoje+"&final_date="+hoje+"&from_status=RECEBIDA&to_status=ENTREGUE",
		loginMecanico.AccessToken, nil, nil,
	)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("esperado 403 pra papel não-admin, veio %d, body %q", rec.Code, rec.Body.String())
	}
}

func TestRelatorioTransicaoStatusRoute_ValidacaoDeQuery_Retorna400(t *testing.T) {
	resetDB(t)
	seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	hoje := time.Now().UTC().Format("2006-01-02")

	casos := []string{
		"start_date=" + hoje + "&final_date=" + hoje + "&from_status=RECEBIDA&to_status=ENTREGUE&unit=dias",
		"final_date=" + hoje + "&from_status=RECEBIDA&to_status=ENTREGUE",
		"start_date=2025-12-31&final_date=" + hoje + "&from_status=RECEBIDA&to_status=ENTREGUE",
		"start_date=" + hoje + "&final_date=" + hoje + "&from_status=APROVADA&to_status=RECEBIDA",
	}

	for _, query := range casos {
		rec := doRequest(t, http.MethodGet,
			"/v1/relatorios/ordens-servico/transicao-status?"+query,
			loginAdmin.AccessToken, nil, nil,
		)
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("query %q: esperado 400, veio %d, body %q", query, rec.Code, rec.Body.String())
		}
	}
}
