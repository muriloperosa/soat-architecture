//go:build integration

package integration_test

import (
	"context"
	"net/http"
	"strconv"
	"testing"

	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
)

func TestOrcamento_GerarAdicionarERemoverItens_CalculaTotais(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")
	ordemServicoID := seedOrdemServico(t, admin.ID)

	servico, err := testContainer.CriarServicoUC.Executar(context.Background(), appservico.CriarServicoInput{
		Nome:                 "Troca de óleo",
		Descricao:            "Troca de óleo do motor",
		PrecoBase:            100.0,
		TempoEstimadoMinutos: 60,
		CriadoPor:            admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar serviço: %v", err)
	}

	peca, err := testContainer.CadastrarPecaUC.Executar(context.Background(), apppeca.CadastrarPecaInput{
		Nome:                "Filtro de óleo",
		Marca:               "Marca X",
		Descricao:           "Filtro de óleo do motor",
		Preco:               50.0,
		QuantidadeEmEstoque: 10,
		EstoqueMinimo:       2,
		CriadoPor:           admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao cadastrar peça: %v", err)
	}

	osPath := "/v1/ordens-servico/" + strconv.FormatUint(ordemServicoID, 10)

	var gerado httporcamento.OrcamentoResponse
	rec := doRequest(t, http.MethodPost, osPath+"/orcamento", login.AccessToken,
		httporcamento.GerarOrcamentoRequest{Observacoes: "Aguardando aprovação"}, &gerado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gerar orçamento: status %d, body %q", rec.Code, rec.Body.String())
	}
	if gerado.OrdemServicoID != ordemServicoID {
		t.Fatalf("ordem de serviço inesperada: %d", gerado.OrdemServicoID)
	}

	rec = doRequest(t, http.MethodPost, osPath+"/orcamento", login.AccessToken,
		httporcamento.GerarOrcamentoRequest{}, nil)
	if rec.Code != http.StatusConflict {
		t.Fatalf("gerar orçamento duplicado deveria retornar 409, veio %d", rec.Code)
	}

	var comServico httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento/itens-servico", login.AccessToken,
		httporcamento.AdicionarServicoOrcamentoRequest{ServicoID: servico.ID, Quantidade: 2}, &comServico)
	if rec.Code != http.StatusOK {
		t.Fatalf("adicionar item de serviço: status %d, body %q", rec.Code, rec.Body.String())
	}
	if comServico.ValorItemServicos != 200.0 || comServico.ValorTotal != 200.0 {
		t.Fatalf("totais inesperados após adicionar serviço: %+v", comServico)
	}

	var comPeca httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento/itens-peca", login.AccessToken,
		httporcamento.AdicionarPecaOrcamentoRequest{PecaID: peca.ID, Quantidade: 3}, &comPeca)
	if rec.Code != http.StatusOK {
		t.Fatalf("adicionar item de peça: status %d, body %q", rec.Code, rec.Body.String())
	}
	if comPeca.ValorItemPecas != 150.0 || comPeca.ValorTotal != 350.0 {
		t.Fatalf("totais inesperados após adicionar peça: %+v", comPeca)
	}
	if len(comPeca.ItensServico) != 1 || len(comPeca.ItensPeca) != 1 {
		t.Fatalf("itens inesperados: %+v", comPeca)
	}

	var totalPersistido float64
	if err := testDB.Table("orcamentos").
		Select("valor_total").
		Where("id = ?", gerado.ID).
		Scan(&totalPersistido).Error; err != nil {
		t.Fatalf("erro ao consultar total persistido: %v", err)
	}
	if totalPersistido != 350.0 {
		t.Fatalf("total persistido inesperado: %v", totalPersistido)
	}

	// Alterar o serviço no catálogo depois da inclusão não deve afetar o
	// item já incluído no orçamento (valor histórico preservado).
	_, err = testContainer.AtualizarServicoUC.Executar(context.Background(), appservico.AtualizarServicoInput{
		ID:                   servico.ID,
		Nome:                 servico.Nome,
		Descricao:            servico.Descricao,
		PrecoBase:            999.0,
		TempoEstimadoMinutos: 60,
	})
	if err != nil {
		t.Fatalf("erro ao atualizar serviço: %v", err)
	}

	itemServicoID := comPeca.ItensServico[0].ID
	var semServico httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodDelete, osPath+"/orcamento/itens-servico/"+strconv.FormatUint(itemServicoID, 10),
		login.AccessToken, nil, &semServico)
	if rec.Code != http.StatusOK {
		t.Fatalf("remover item de serviço: status %d, body %q", rec.Code, rec.Body.String())
	}
	if len(semServico.ItensServico) != 0 {
		t.Fatalf("item de serviço deveria ter sido removido: %+v", semServico)
	}
	if semServico.ValorTotal != 150.0 {
		t.Fatalf("total inesperado após remover serviço (deveria preservar o preço original, não o atualizado): %+v", semServico)
	}

	itemPecaID := comPeca.ItensPeca[0].ID
	var semPeca httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodDelete, osPath+"/orcamento/itens-peca/"+strconv.FormatUint(itemPecaID, 10),
		login.AccessToken, nil, &semPeca)
	if rec.Code != http.StatusOK {
		t.Fatalf("remover item de peça: status %d, body %q", rec.Code, rec.Body.String())
	}
	if semPeca.ValorTotal != 0 {
		t.Fatalf("total deveria zerar após remover todos os itens: %+v", semPeca)
	}

	rec = doRequest(t, http.MethodDelete, osPath+"/orcamento/itens-peca/"+strconv.FormatUint(itemPecaID, 10),
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("remover item de peça já removido deveria retornar 404, veio %d", rec.Code)
	}
}

func TestOrcamento_Finalizar_TransicionaOSEEnviaEmailAoCliente(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")
	ordemServicoID := seedOrdemServico(t, admin.ID)

	servico, err := testContainer.CriarServicoUC.Executar(context.Background(), appservico.CriarServicoInput{
		Nome:                 "Troca de óleo",
		Descricao:            "Troca de óleo do motor",
		PrecoBase:            100.0,
		TempoEstimadoMinutos: 60,
		CriadoPor:            admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar serviço: %v", err)
	}

	osPath := "/v1/ordens-servico/" + strconv.FormatUint(ordemServicoID, 10)

	var gerado httporcamento.OrcamentoResponse
	rec := doRequest(t, http.MethodPost, osPath+"/orcamento", login.AccessToken, httporcamento.GerarOrcamentoRequest{}, &gerado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gerar orçamento: status %d, body %q", rec.Code, rec.Body.String())
	}

	var comServico httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento/itens-servico", login.AccessToken,
		httporcamento.AdicionarServicoOrcamentoRequest{ServicoID: servico.ID, Quantidade: 2}, &comServico)
	if rec.Code != http.StatusOK {
		t.Fatalf("adicionar item de serviço: status %d, body %q", rec.Code, rec.Body.String())
	}

	// A OS precisa estar EM_DIAGNOSTICO com diagnóstico preenchido antes de
	// finalizar o orçamento (mesma invariante de transição da OS).
	rec = doRequest(t, http.MethodPatch, "/v1/ordens-servico/"+strconv.FormatUint(ordemServicoID, 10)+"/iniciar-diagnostico",
		login.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, http.MethodPut, "/v1/ordens-servico/"+strconv.FormatUint(ordemServicoID, 10)+"/diagnostico",
		login.AccessToken, httpordemservico.InformarDiagnosticoRequest{Diagnostico: "Falha na bomba de combustível"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("informar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}

	var finalizado httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPatch, osPath+"/orcamento/finalizar", login.AccessToken, nil, &finalizado)
	if rec.Code != http.StatusOK {
		t.Fatalf("finalizar orçamento: status %d, body %q", rec.Code, rec.Body.String())
	}
	if finalizado.ValorTotal != 200.0 {
		t.Fatalf("total inesperado ao finalizar: %+v", finalizado)
	}

	var statusOS string
	if err := testDB.Table("ordens_servico").Select("status").Where("id = ?", ordemServicoID).Scan(&statusOS).Error; err != nil {
		t.Fatalf("erro ao consultar status da OS: %v", err)
	}
	if statusOS != "AGUARDANDO_APROVACAO" {
		t.Fatalf("status da OS inesperado após finalizar orçamento: %q", statusOS)
	}

	rec = doRequest(t, http.MethodPatch, "/v1/ordens-servico/999999/orcamento/finalizar", login.AccessToken, nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("finalizar orçamento de OS inexistente deveria retornar 404, veio %d", rec.Code)
	}
}
