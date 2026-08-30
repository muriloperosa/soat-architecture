//go:build integration

package integration_test

import (
	"fmt"
	"net/http"
	"testing"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httporcamento "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/orcamento"
	httpordemservico "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/ordemservico"
)

func TestConsultaOrdemServico_ListarBuscarPorIDEPorNumero(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	cliente, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725", Tipo: string(domaincliente.TipoPessoaFisica), Nome: "Maria Silva",
		Email: "maria.consulta@email.com", Telefone: "11999998888", Senha: "senhaCliente123", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}

	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "ABC1D23", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 52000,
		Ano: 2020, Cor: "Prata", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	var aberta httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", login.AccessToken, httpordemservico.AbrirOrdemServicoRequest{
		ClienteID: cliente.ID, VeiculoID: veiculo.ID, QuilometragemEntrada: 52300, Observacoes: "Ruído no motor",
	}, &aberta)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS: status %d, body %q", rec.Code, rec.Body.String())
	}

	var listagem httpordemservico.ListarOrdensServicoResponse
	path := fmt.Sprintf("/v1/ordens-servico?page=1&status=RECEBIDA&cliente_id=%d&veiculo_id=%d", cliente.ID, veiculo.ID)
	rec = doRequest(t, http.MethodGet, path, login.AccessToken, nil, &listagem)
	if rec.Code != http.StatusOK {
		t.Fatalf("listar OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if listagem.Total != 1 || len(listagem.Items) != 1 {
		t.Fatalf("listagem inesperada: %+v", listagem)
	}
	if listagem.Page != 1 || listagem.PageSize != 20 || listagem.TotalPages != 1 {
		t.Fatalf("paginação inesperada: %+v", listagem)
	}
	if listagem.Items[0].ID != aberta.ID || listagem.Items[0].Status != "RECEBIDA" {
		t.Fatalf("OS inesperada na listagem: %+v", listagem.Items[0])
	}
	if listagem.Items[0].Orcamento != nil {
		t.Fatalf("OS recém-aberta não deveria possuir orçamento: %+v", listagem.Items[0].Orcamento)
	}

	var porID httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/ordens-servico/%d", aberta.ID), login.AccessToken, nil, &porID)
	if rec.Code != http.StatusOK {
		t.Fatalf("buscar OS por ID: status %d, body %q", rec.Code, rec.Body.String())
	}
	if porID.ID != aberta.ID || len(porID.HistoricoStatus) != 1 {
		t.Fatalf("consulta por ID inesperada: %+v", porID)
	}

	var porNumero httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico/numero/"+aberta.Numero, login.AccessToken, nil, &porNumero)
	if rec.Code != http.StatusOK {
		t.Fatalf("buscar OS por número: status %d, body %q", rec.Code, rec.Body.String())
	}
	if porNumero.ID != aberta.ID || porNumero.Numero != aberta.Numero {
		t.Fatalf("consulta por número inesperada: %+v", porNumero)
	}
}

func TestConsultaOrdemServico_ListagemRetornaResumoDoOrcamento(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")

	cliente, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Maria Orçamento",
		Email:     "maria.orcamento@email.com",
		Telefone:  "11999998888",
		Senha:     "senhaCliente123",
		CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente: %v", err)
	}

	veiculo, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa:              "ORC1A23",
		Marca:              "Honda",
		Modelo:             "Civic",
		QuilometragemAtual: 30000,
		Ano:                2022,
		Cor:                "Preto",
		CriadoPor:          admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo: %v", err)
	}

	var aberta httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", login.AccessToken, httpordemservico.AbrirOrdemServicoRequest{
		ClienteID:            cliente.ID,
		VeiculoID:            veiculo.ID,
		QuilometragemEntrada: 30100,
	}, &aberta)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS: status %d, body %q", rec.Code, rec.Body.String())
	}

	osPath := fmt.Sprintf("/v1/ordens-servico/%d", aberta.ID)
	rec = doRequest(t, http.MethodPatch, osPath+"/iniciar-diagnostico", login.AccessToken, nil, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("iniciar diagnóstico: status %d, body %q", rec.Code, rec.Body.String())
	}

	var orcamento httporcamento.OrcamentoResponse
	rec = doRequest(t, http.MethodPost, osPath+"/orcamento", login.AccessToken,
		httporcamento.GerarOrcamentoRequest{Observacoes: "Orçamento inicial"}, &orcamento)
	if rec.Code != http.StatusCreated {
		t.Fatalf("gerar orçamento: status %d, body %q", rec.Code, rec.Body.String())
	}

	var listagem httpordemservico.ListarOrdensServicoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/ordens-servico?cliente_id=%d", cliente.ID), login.AccessToken, nil, &listagem)
	if rec.Code != http.StatusOK {
		t.Fatalf("listar OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if len(listagem.Items) != 1 {
		t.Fatalf("esperada uma OS, encontrado %+v", listagem)
	}

	resumo := listagem.Items[0].Orcamento
	if resumo == nil {
		t.Fatal("listagem deveria retornar o resumo do orçamento")
	}
	if resumo.ID != orcamento.ID || resumo.ValorTotal != orcamento.ValorTotal || resumo.Observacoes != "Orçamento inicial" {
		t.Fatalf("resumo do orçamento inesperado: %+v", resumo)
	}
}

func TestConsultaOrdemServico_ClienteListaSomenteAsPropriasOrdens(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	clienteA, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Cliente A",
		Email:     "cliente.a@email.com",
		Telefone:  "11999998888",
		Senha:     "senhaClienteA123",
		CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente A: %v", err)
	}

	clienteB, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "11144477735",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Cliente B",
		Email:     "cliente.b@email.com",
		Telefone:  "11988887777",
		Senha:     "senhaClienteB123",
		CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente B: %v", err)
	}

	veiculoA, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "AAA1A11", Marca: "Fiat", Modelo: "Uno", QuilometragemAtual: 10000,
		Ano: 2020, Cor: "Prata", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo A: %v", err)
	}

	veiculoB, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "BBB2B22", Marca: "Ford", Modelo: "Ka", QuilometragemAtual: 20000,
		Ano: 2021, Cor: "Branco", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo B: %v", err)
	}

	var osA httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", loginAdmin.AccessToken, httpordemservico.AbrirOrdemServicoRequest{
		ClienteID: clienteA.ID, VeiculoID: veiculoA.ID, QuilometragemEntrada: 10100,
	}, &osA)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS do cliente A: status %d, body %q", rec.Code, rec.Body.String())
	}

	var osB httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodPost, "/v1/ordens-servico", loginAdmin.AccessToken, httpordemservico.AbrirOrdemServicoRequest{
		ClienteID: clienteB.ID, VeiculoID: veiculoB.ID, QuilometragemEntrada: 20100,
	}, &osB)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS do cliente B: status %d, body %q", rec.Code, rec.Body.String())
	}

	loginClienteA := doLoginCliente(t, "cliente.a@email.com", "senhaClienteA123")

	var listagem httpordemservico.ListarOrdensServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico", loginClienteA.AccessToken, nil, &listagem)
	if rec.Code != http.StatusOK {
		t.Fatalf("cliente listar próprias OS: status %d, body %q", rec.Code, rec.Body.String())
	}
	if listagem.Total != 1 || len(listagem.Items) != 1 || listagem.Items[0].ID != osA.ID {
		t.Fatalf("cliente deveria visualizar somente a própria OS: %+v", listagem)
	}
	if listagem.Items[0].ClienteID != clienteA.ID {
		t.Fatalf("OS retornada pertence a outro cliente: %+v", listagem.Items[0])
	}

	// Mesmo tentando filtrar explicitamente pelo outro cliente, a restrição
	// derivada do JWT continua sendo aplicada e impede vazamento de dados.
	var tentativaOutroCliente httpordemservico.ListarOrdensServicoResponse
	rec = doRequest(t, http.MethodGet,
		fmt.Sprintf("/v1/ordens-servico?cliente_id=%d", clienteB.ID),
		loginClienteA.AccessToken, nil, &tentativaOutroCliente)
	if rec.Code != http.StatusOK {
		t.Fatalf("filtro por outro cliente deveria resultar em lista vazia, status %d, body %q", rec.Code, rec.Body.String())
	}
	if tentativaOutroCliente.Total != 0 || len(tentativaOutroCliente.Items) != 0 {
		t.Fatalf("cliente não deveria conseguir visualizar OS alheia via filtro: %+v", tentativaOutroCliente)
	}
}

func TestConsultaOrdemServico_ClienteBuscaSomenteAsPropriasOrdensPorIDEPorNumero(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	loginAdmin := doLogin(t, "admin@oficina.com", "senha123")

	clienteA, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "52998224725",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Cliente A",
		Email:     "cliente.a.detalhe@email.com",
		Telefone:  "11999998888",
		Senha:     "senhaClienteA123",
		CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente A: %v", err)
	}

	clienteB, err := testContainer.CriarClienteUseCase.Executar(t.Context(), appcliente.CriarClienteInput{
		Documento: "11144477735",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Cliente B",
		Email:     "cliente.b.detalhe@email.com",
		Telefone:  "11988887777",
		Senha:     "senhaClienteB123",
		CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar cliente B: %v", err)
	}

	veiculoA, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "CCC3C33", Marca: "Honda", Modelo: "Fit", QuilometragemAtual: 30000,
		Ano: 2020, Cor: "Cinza", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo A: %v", err)
	}

	veiculoB, err := testContainer.CadastrarVeiculoUC.Executar(t.Context(), appveiculo.CadastrarVeiculoInput{
		Placa: "DDD4D44", Marca: "Toyota", Modelo: "Etios", QuilometragemAtual: 40000,
		Ano: 2021, Cor: "Preto", CriadoPor: admin.ID,
	})
	if err != nil {
		t.Fatalf("erro ao criar veículo B: %v", err)
	}

	var osA httpordemservico.OrdemServicoResponse
	rec := doRequest(t, http.MethodPost, "/v1/ordens-servico", loginAdmin.AccessToken, httpordemservico.AbrirOrdemServicoRequest{
		ClienteID: clienteA.ID, VeiculoID: veiculoA.ID, QuilometragemEntrada: 30100,
	}, &osA)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS do cliente A: status %d, body %q", rec.Code, rec.Body.String())
	}

	var osB httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodPost, "/v1/ordens-servico", loginAdmin.AccessToken, httpordemservico.AbrirOrdemServicoRequest{
		ClienteID: clienteB.ID, VeiculoID: veiculoB.ID, QuilometragemEntrada: 40100,
	}, &osB)
	if rec.Code != http.StatusCreated {
		t.Fatalf("abrir OS do cliente B: status %d, body %q", rec.Code, rec.Body.String())
	}

	loginClienteA := doLoginCliente(t, "cliente.a.detalhe@email.com", "senhaClienteA123")

	var porID httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/ordens-servico/%d", osA.ID), loginClienteA.AccessToken, nil, &porID)
	if rec.Code != http.StatusOK || porID.ID != osA.ID {
		t.Fatalf("cliente deveria consultar a própria OS por ID: status %d, body %q", rec.Code, rec.Body.String())
	}

	var porNumero httpordemservico.OrdemServicoResponse
	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico/numero/"+osA.Numero, loginClienteA.AccessToken, nil, &porNumero)
	if rec.Code != http.StatusOK || porNumero.ID != osA.ID {
		t.Fatalf("cliente deveria consultar a própria OS por número: status %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodGet, fmt.Sprintf("/v1/ordens-servico/%d", osB.ID), loginClienteA.AccessToken, nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cliente consultando OS alheia por ID deveria receber 403, veio %d, body %q", rec.Code, rec.Body.String())
	}

	rec = doRequest(t, http.MethodGet, "/v1/ordens-servico/numero/"+osB.Numero, loginClienteA.AccessToken, nil, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cliente consultando OS alheia por número deveria receber 403, veio %d, body %q", rec.Code, rec.Body.String())
	}
}
