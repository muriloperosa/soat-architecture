//go:build integration

package integration_test

import (
	"net/http"
	"strconv"
	"testing"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
)

func TestDiagnosticoOrdemServico_PersisteEstadoHistoricoETexto(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	mecanico := seedUsuario(t, "Mecânico Oficina", "mecanico@oficina.com", "senha123", shared.PapelMecanico)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")
	loginMecanico := doLogin(t, "mecanico@oficina.com", "senha123")

	cliente, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Maria Silva",
		Email:     "maria@email.com",
		Telefone:  "11999998888",
		Senha:     "senhaCliente123",
		CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}

	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa:              "ABC1D23",
		Marca:              "Fiat",
		Modelo:             "Uno",
		QuilometragemAtual: 52_000,
		Ano:                2020,
		Cor:                "Prata",
		CriadoPor:          admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	var aberta httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", loginAdmin.AccessToken,
		httpordemservico.AbrirOrdemServicoRequest{
			ClienteID:            cliente.ID,
			VeiculoID:            veiculo.ID,
			QuilometragemEntrada: 52_300,
		},
		&aberta,
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS: status %d, body %q", rec.Code, rec.Body.String())
	}

	var iniciada httpordemservico.OrdemServicoResponse
	rec = doRequest(
		t,
		http.MethodPatch,
		"/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/iniciar-diagnostico",
		loginMecanico.AccessToken,
		nil,
		&iniciada,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}
	if iniciada.Status != "EM_DIAGNOSTICO" {
		t.Fatalf("status inesperado: %s", iniciada.Status)
	}
	if iniciada.Diagnostico != "" {
		t.Fatalf("diagnóstico deve iniciar vazio, recebido %q", iniciada.Diagnostico)
	}
	if len(iniciada.HistoricoStatus) != 2 {
		t.Fatalf("esperados dois históricos na resposta, encontrado %d", len(iniciada.HistoricoStatus))
	}
	for _, h := range iniciada.HistoricoStatus {
		if h.ID == 0 {
			t.Fatalf("histórico deveria vir com ID real do banco, veio: %+v", iniciada.HistoricoStatus)
		}
	}

	var quantidadeHistoricos int64
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ?", aberta.ID).
		Count(&quantidadeHistoricos).Error; err != nil {
		t.Fatalf("erro ao consultar históricos: %v", err)
	}
	if quantidadeHistoricos != 2 {
		t.Fatalf("esperados dois históricos, encontrado %d", quantidadeHistoricos)
	}

	var historico struct {
		Status      string
		AlteradoPor uint64
		Motivo      string
	}
	if err := testDB.Table("historicos_status").
		Where("ordem_servico_id = ? AND status = ?", aberta.ID, "EM_DIAGNOSTICO").
		Take(&historico).Error; err != nil {
		t.Fatalf("erro ao consultar histórico de diagnóstico: %v", err)
	}
	if historico.AlteradoPor != mecanico.ID || historico.Motivo != "" {
		t.Fatalf("histórico inesperado: %+v", historico)
	}

	var diagnosticada httpordemservico.OrdemServicoResponse
	rec = doRequest(
		t,
		http.MethodPut,
		"/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/diagnostico",
		loginMecanico.AccessToken,
		httpordemservico.InformarDiagnosticoRequest{Diagnostico: "Falha na bomba de combustível"},
		&diagnosticada,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf("informar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}
	if diagnosticada.Diagnostico != "Falha na bomba de combustível" {
		t.Fatalf("diagnóstico inesperado: %q", diagnosticada.Diagnostico)
	}

	var diagnosticoPersistido string
	if err := testDB.Table("ordens_servico").
		Select("diagnostico").
		Where("id = ?", aberta.ID).
		Scan(&diagnosticoPersistido).Error; err != nil {
		t.Fatalf("erro ao consultar diagnóstico persistido: %v", err)
	}
	if diagnosticoPersistido != "Falha na bomba de combustível" {
		t.Fatalf("diagnóstico não persistido: %q", diagnosticoPersistido)
	}
}

func TestDiagnosticoOrdemServico_ValidaTransicaoETextoVazio(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	cliente, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725", Tipo: string(domaincliente.TipoPessoaFisica), Nome: "Maria Silva",
		Email: "maria@email.com", Telefone: "11999998888", Senha: "senhaCliente123", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}
	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 52_000,
		Ano: 2020, Cor: "Prata", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	var aberta httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", login.AccessToken,
		httpordemservico.AbrirOrdemServicoRequest{ClienteID: cliente.ID, VeiculoID: veiculo.ID}, &aberta)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPut, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/diagnostico",
		login.AccessToken, httpordemservico.InformarDiagnosticoRequest{Diagnostico: "Falha na bomba"}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("diagnóstico antes da transição deveria retornar 400, veio %d", rec.Code)
	}

	rec = doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/iniciar-diagnostico",
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/iniciar-diagnostico",
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("repetir transição deveria retornar 400, veio %d", rec.Code)
	}

	rec = doRequest(t, http.MethodPut, "/v1/ordens-servico/"+strconv.FormatUint(aberta.ID, 10)+"/diagnostico",
		login.AccessToken, httpordemservico.InformarDiagnosticoRequest{Diagnostico: "   "}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("diagnóstico vazio deveria retornar 400, veio %d", rec.Code)
	}
}
