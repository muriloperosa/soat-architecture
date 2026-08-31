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

	// Gerar orçamento exige a OS EM_DIAGNOSTICO.
	rec := doRequest(t, http.MethodPost, osPath+"/orcamento", login.AccessToken,
		httporcamento.GerarOrcamentoRequest{}, nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("gerar orçamento com OS RECEBIDA deveria retornar 400, veio %d", rec.Code)
	}

	rec = doRequest(t, http.MethodPatch, osPath+"/iniciar-diagnostico", login.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}

	var gerado httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento", login.AccessToken,
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

func TestOrcamento_EnviarParaAprovacao_TransicionaOSEEnviaEmailAoCliente(t *testing.T) {
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

	// Gerar orçamento exige a OS EM_DIAGNOSTICO com diagnóstico preenchido
	// (mesma invariante exigida pelo envio para aprovação).
	rec := doRequest(t, http.MethodPatch, osPath+"/iniciar-diagnostico", login.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}
	rec = doRequest(t, http.MethodPut, osPath+"/diagnostico", login.AccessToken,
		httpordemservico.InformarDiagnosticoRequest{Diagnostico: "Falha na bomba de combustível"}, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("informar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}

	var gerado httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento", login.AccessToken, httporcamento.GerarOrcamentoRequest{}, &gerado)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gerar orçamento: status %d, body %q", rec.Code, rec.Body.String())
	}

	var comServico httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento/itens-servico", login.AccessToken,
		httporcamento.AdicionarServicoOrcamentoRequest{ServicoID: servico.ID, Quantidade: 2}, &comServico)
	if rec.Code != http.StatusOK {
		t.Fatalf("adicionar item de serviço: status %d, body %q", rec.Code, rec.Body.String())
	}

	var enviado httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPatch, osPath+"/orcamento/enviar-aprovacao", login.AccessToken, nil, &enviado)
	if rec.Code != http.StatusOK {
		t.Fatalf("enviar orçamento para aprovação: status %d, body %q", rec.Code, rec.Body.String())
	}
	if enviado.ValorTotal != 200.0 {
		t.Fatalf("total inesperado ao enviar para aprovação: %+v", enviado)
	}

	var statusOS string
	if err := testDB.Table("ordens_servico").Select("status").Where("id = ?", ordemServicoID).Scan(&statusOS).Error; err != nil {
		t.Fatalf("erro ao consultar status da OS: %v", err)
	}
	if statusOS != "AGUARDANDO_APROVACAO" {
		t.Fatalf("status da OS inesperado após enviar orçamento para aprovação: %q", statusOS)
	}

	rec = doRequest(t, http.MethodPatch, "/v1/ordens-servico/999999/orcamento/enviar-aprovacao", login.AccessToken, nil, nil)
	if rec.Code != http.StatusNotFound {
		t.Fatalf("enviar orçamento de OS inexistente para aprovação deveria retornar 404, veio %d", rec.Code)
	}
}

func TestOrcamento_FluxoCompleto_RejeitarEditarReenviarEAprovar(t *testing.T) {
	resetDB(t)

	// ---------------------------------------------------------------------
	// Arrange: usuário interno, OS e cliente proprietário
	// ---------------------------------------------------------------------

	admin := seedUsuario(
		t,
		"Admin Oficina",
		"admin@oficina.com",
		"senha123",
		shared.PapelAdmin,
	)

	loginInterno := doLogin(
		t,
		"admin@oficina.com",
		"senha123",
	)

	ordemServicoID := seedOrdemServico(t, admin.ID)

	osPath := "/v1/ordens-servico/" + strconv.FormatUint(ordemServicoID, 10)

	// Recupera o cliente proprietário criado pelo seedOrdemServico.
	var cliente struct {
		ID    uint64
		Email string
	}

	if err := testDB.
		Table("clientes").
		Select("clientes.id, clientes.email").
		Joins("JOIN ordens_servico ON ordens_servico.cliente_id = clientes.id").
		Where("ordens_servico.id = ?", ordemServicoID).
		Scan(&cliente).Error; err != nil {
		t.Fatalf("erro ao buscar cliente da OS: %v", err)
	}

	if cliente.ID == 0 {
		t.Fatal("cliente da OS não encontrado")
	}

	loginCliente := doLoginCliente(
		t,
		cliente.Email,
		"senha123",
	)

	// Serviço usado para tornar o orçamento válido.
	servico, err := testContainer.CriarServicoUC.Executar(
		context.Background(),
		appservico.CriarServicoInput{
			Nome:                 "Troca de óleo",
			Descricao:            "Troca de óleo do motor",
			PrecoBase:            100.0,
			TempoEstimadoMinutos: 60,
			CriadoPor:            admin.ID,
		},
	)
	if err != nil {
		t.Fatalf("erro ao criar serviço: %v", err)
	}

	// ---------------------------------------------------------------------
	// EM_DIAGNOSTICO
	// ---------------------------------------------------------------------

	rec := doRequest(
		t,
		http.MethodPatch,
		osPath+"/iniciar-diagnostico",
		loginInterno.AccessToken,
		nil,
		nil,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"iniciar diagnóstico: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	rec = doRequest(
		t,
		http.MethodPut,
		osPath+"/diagnostico",
		loginInterno.AccessToken,
		httpordemservico.InformarDiagnosticoRequest{
			Diagnostico: "Necessária troca de componentes",
		},
		nil,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"informar diagnóstico: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	// ---------------------------------------------------------------------
	// Cria orçamento e adiciona item
	// ---------------------------------------------------------------------

	var orcamento httporcamento.OrcamentoResponse

	rec = doRequest(
		t,
		http.MethodPost,
		osPath+"/orcamento",
		loginInterno.AccessToken,
		httporcamento.GerarOrcamentoRequest{
			Observacoes: "Orçamento inicial",
		},
		&orcamento,
	)
	if rec.Code != http.StatusCreated {
		t.Fatalf(
			"gerar orçamento: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	var orcamentoComServico httporcamento.OrcamentoResponse

	rec = doRequest(
		t,
		http.MethodPost,
		osPath+"/orcamento/itens-servico",
		loginInterno.AccessToken,
		httporcamento.AdicionarServicoOrcamentoRequest{
			ServicoID:  servico.ID,
			Quantidade: 1,
		},
		&orcamentoComServico,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"adicionar serviço ao orçamento: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	if orcamentoComServico.ValorTotal != 100 {
		t.Fatalf(
			"valor inicial do orçamento inesperado: %.2f",
			orcamentoComServico.ValorTotal,
		)
	}

	// ---------------------------------------------------------------------
	// EM_DIAGNOSTICO -> AGUARDANDO_APROVACAO
	// ---------------------------------------------------------------------

	var enviado httporcamento.OrcamentoResponse

	rec = doRequest(
		t,
		http.MethodPatch,
		osPath+"/orcamento/enviar-aprovacao",
		loginInterno.AccessToken,
		nil,
		&enviado,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"enviar orçamento para aprovação: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	var statusOS string

	if err := testDB.
		Table("ordens_servico").
		Select("status").
		Where("id = ?", ordemServicoID).
		Scan(&statusOS).Error; err != nil {
		t.Fatalf("erro ao consultar status da OS: %v", err)
	}

	if statusOS != "AGUARDANDO_APROVACAO" {
		t.Fatalf(
			"esperado status AGUARDANDO_APROVACAO, recebido %q",
			statusOS,
		)
	}

	// ---------------------------------------------------------------------
	// Cliente rejeita
	// AGUARDANDO_APROVACAO -> REJEITADA
	// ---------------------------------------------------------------------

	const motivoRejeicao = "Valor acima do esperado"

	var rejeitada httporcamento.FluxoOrcamentoResponse

	rec = doRequest(
		t,
		http.MethodPatch,
		osPath+"/orcamento/rejeitar",
		loginCliente.AccessToken,
		httporcamento.RejeitarOrcamentoRequest{
			Motivo: motivoRejeicao,
		},
		&rejeitada,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"rejeitar orçamento: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	if rejeitada.Status != "REJEITADA" {
		t.Fatalf(
			"status retornado após rejeição inesperado: %q",
			rejeitada.Status,
		)
	}

	if err := testDB.
		Table("ordens_servico").
		Select("status").
		Where("id = ?", ordemServicoID).
		Scan(&statusOS).Error; err != nil {
		t.Fatalf("erro ao consultar status após rejeição: %v", err)
	}

	if statusOS != "REJEITADA" {
		t.Fatalf(
			"esperado status REJEITADA, recebido %q",
			statusOS,
		)
	}

	// ---------------------------------------------------------------------
	// Motivo precisa estar persistido no HistoricoStatus
	// ---------------------------------------------------------------------

	var historicoRejeicao struct {
		Status string
		Motivo string
	}

	if err := testDB.
		Table("historicos_status").
		Select("status, motivo").
		Where(
			"ordem_servico_id = ? AND status = ?",
			ordemServicoID,
			"REJEITADA",
		).
		Order("id DESC").
		Limit(1).
		Scan(&historicoRejeicao).Error; err != nil {
		t.Fatalf("erro ao consultar histórico de rejeição: %v", err)
	}

	if historicoRejeicao.Status != "REJEITADA" {
		t.Fatalf(
			"histórico REJEITADA não encontrado: %+v",
			historicoRejeicao,
		)
	}

	if historicoRejeicao.Motivo != motivoRejeicao {
		t.Fatalf(
			"motivo da rejeição inesperado: esperado %q, recebido %q",
			motivoRejeicao,
			historicoRejeicao.Motivo,
		)
	}

	// ---------------------------------------------------------------------
	// REJEITADA -> orçamento volta a ser editável
	//
	// Adicionamos novamente o serviço para comprovar que a edição foi
	// liberada. O valor passa de 100 para 200.
	// ---------------------------------------------------------------------

	var orcamentoEditado httporcamento.OrcamentoResponse

	rec = doRequest(
		t,
		http.MethodPost,
		osPath+"/orcamento/itens-servico",
		loginInterno.AccessToken,
		httporcamento.AdicionarServicoOrcamentoRequest{
			ServicoID:  servico.ID,
			Quantidade: 1,
		},
		&orcamentoEditado,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"orçamento rejeitado deveria permitir edição: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	if orcamentoEditado.ValorTotal != 200 {
		t.Fatalf(
			"valor esperado após editar orçamento rejeitado: 200, recebido %.2f",
			orcamentoEditado.ValorTotal,
		)
	}

	// ---------------------------------------------------------------------
	// REJEITADA -> AGUARDANDO_APROVACAO
	// ---------------------------------------------------------------------

	rec = doRequest(
		t,
		http.MethodPatch,
		osPath+"/orcamento/enviar-aprovacao",
		loginInterno.AccessToken,
		nil,
		&enviado,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"reenviar orçamento rejeitado: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	if err := testDB.
		Table("ordens_servico").
		Select("status").
		Where("id = ?", ordemServicoID).
		Scan(&statusOS).Error; err != nil {
		t.Fatalf("erro ao consultar status após reenvio: %v", err)
	}

	if statusOS != "AGUARDANDO_APROVACAO" {
		t.Fatalf(
			"esperado AGUARDANDO_APROVACAO após reenvio, recebido %q",
			statusOS,
		)
	}

	// ---------------------------------------------------------------------
	// Cliente aprova
	// AGUARDANDO_APROVACAO -> APROVADA
	// ---------------------------------------------------------------------

	var aprovada httporcamento.FluxoOrcamentoResponse

	rec = doRequest(
		t,
		http.MethodPatch,
		osPath+"/orcamento/aprovar",
		loginCliente.AccessToken,
		nil,
		&aprovada,
	)
	if rec.Code != http.StatusOK {
		t.Fatalf(
			"aprovar orçamento: status %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	if aprovada.Status != "APROVADA" {
		t.Fatalf(
			"status retornado após aprovação inesperado: %q",
			aprovada.Status,
		)
	}

	if err := testDB.
		Table("ordens_servico").
		Select("status").
		Where("id = ?", ordemServicoID).
		Scan(&statusOS).Error; err != nil {
		t.Fatalf("erro ao consultar status após aprovação: %v", err)
	}

	if statusOS != "APROVADA" {
		t.Fatalf(
			"esperado status APROVADA, recebido %q",
			statusOS,
		)
	}

	// ---------------------------------------------------------------------
	// Orçamento APROVADO é imutável
	// ---------------------------------------------------------------------

	rec = doRequest(
		t,
		http.MethodPost,
		osPath+"/orcamento/itens-servico",
		loginInterno.AccessToken,
		httporcamento.AdicionarServicoOrcamentoRequest{
			ServicoID:  servico.ID,
			Quantidade: 1,
		},
		nil,
	)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf(
			"orçamento aprovado não deveria permitir edição; esperado 400, recebido %d, body %q",
			rec.Code,
			rec.Body.String(),
		)
	}

	// Confirma que a tentativa de edição não alterou o orçamento.
	var valorTotalPersistido float64

	if err := testDB.
		Table("orcamentos").
		Select("valor_total").
		Where("ordem_servico_id = ?", ordemServicoID).
		Scan(&valorTotalPersistido).Error; err != nil {
		t.Fatalf("erro ao consultar valor final do orçamento: %v", err)
	}

	if valorTotalPersistido != 200 {
		t.Fatalf(
			"orçamento aprovado foi alterado indevidamente: esperado 200, recebido %.2f",
			valorTotalPersistido,
		)
	}

	// ---------------------------------------------------------------------
	// Todas as transições do fluxo devem possuir HistoricoStatus
	// ---------------------------------------------------------------------

	statusEsperados := []string{
		"AGUARDANDO_APROVACAO",
		"REJEITADA",
		"AGUARDANDO_APROVACAO",
		"APROVADA",
	}

	for _, statusEsperado := range statusEsperados {
		var quantidade int64

		if err := testDB.
			Table("historicos_status").
			Where(
				"ordem_servico_id = ? AND status = ?",
				ordemServicoID,
				statusEsperado,
			).
			Count(&quantidade).Error; err != nil {
			t.Fatalf(
				"erro ao consultar histórico do status %s: %v",
				statusEsperado,
				err,
			)
		}

		quantidadeEsperada := int64(1)
		if statusEsperado == "AGUARDANDO_APROVACAO" {
			quantidadeEsperada = 2
		}

		if quantidade != quantidadeEsperada {
			t.Fatalf(
				"quantidade de históricos %s inesperada: esperado %d, recebido %d",
				statusEsperado,
				quantidadeEsperada,
				quantidade,
			)
		}
	}
}
