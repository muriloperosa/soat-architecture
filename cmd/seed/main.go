package main

import (
	"context"
	"fmt"
	"log"
	"time"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appordemservico "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	appservico "github.com/muriloperosa/soat-architecture/internal/application/servico"
	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	appveiculo "github.com/muriloperosa/soat-architecture/internal/application/veiculo"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/config"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/persistence/mysql"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/wiring"
)

// senhaPadrao é a senha usada em todos os usuários de login criados pelo
// seed. Mantida fixa e simples de propósito: é só pra ambiente de avaliação.
const senhaPadrao = "Senha@123"

type credencial struct {
	papel string
	nome  string
	email string
	senha string
}

// seed popula o banco com os 4 usuários de login exigidos pra avaliação
// (admin, atendente, mecânico, cliente) e um conjunto pequeno de dados
// fictícios (clientes, veículos, serviços, peças e uma ordem de serviço em
// andamento), rodando direto contra os use cases via wiring.Container, sem
// passar por HTTP/JWT (mesmo approach do cmd/create-user).
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("erro ao carregar config: %v", err)
	}

	db, err := mysql.NewMigrationConnection(cfg)
	if err != nil {
		log.Fatalf("erro ao conectar no banco: %v", err)
	}

	c := wiring.NewContainer(cfg, db)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	credenciais := make([]credencial, 0, 4)

	admin := criarUsuarioInterno(ctx, c, "Ana Administradora", "admin@oficina.com", "ADMINISTRADOR")
	credenciais = append(credenciais, credencial{"admin", admin.Nome, admin.Email, senhaPadrao})

	atendente := criarUsuarioInterno(ctx, c, "Bruno Atendente", "atendente@oficina.com", "ATENDENTE")
	credenciais = append(credenciais, credencial{"atendente", atendente.Nome, atendente.Email, senhaPadrao})

	mecanico := criarUsuarioInterno(ctx, c, "Carlos Mecanico", "mecanico@oficina.com", "MECANICO")
	credenciais = append(credenciais, credencial{"mecanico", mecanico.Nome, mecanico.Email, senhaPadrao})

	clientePrincipal := criarCliente(ctx, c, "52998224725", "Daniela Cliente", "cliente@oficina.com", "11988887777", admin.ID)
	credenciais = append(credenciais, credencial{"cliente", clientePrincipal.Nome, clientePrincipal.Email, senhaPadrao})

	criarCliente(ctx, c, "39053344705", "Eduardo Fake", "eduardo.fake@oficina.com", "11977776666", admin.ID)
	criarCliente(ctx, c, "88221472000", "Fabiana Fake", "fabiana.fake@oficina.com", "11966665555", admin.ID)

	veiculo1 := criarVeiculo(ctx, c, "ABC1234", "Volkswagen", "Gol", 45000, 2018, "Prata", admin.ID)
	criarVeiculo(ctx, c, "XYZ9A87", "Fiat", "Argo", 12000, 2022, "Branco", admin.ID)

	criarServico(ctx, c, "Troca de óleo", "Troca de óleo e filtro", 120.0, 40, admin.ID)
	criarServico(ctx, c, "Alinhamento e balanceamento", "Alinhamento e balanceamento das quatro rodas", 180.0, 60, admin.ID)
	criarServico(ctx, c, "Revisão completa", "Revisão geral com checklist de itens de segurança", 450.0, 180, admin.ID)

	criarPeca(ctx, c, "Filtro de óleo", "Bosch", "Filtro de óleo para motores 1.0 a 2.0", 45.9, 20, 5, admin.ID)
	criarPeca(ctx, c, "Pastilha de freio", "Fras-le", "Jogo de pastilhas de freio dianteiras", 129.9, 15, 4, admin.ID)
	criarPeca(ctx, c, "Correia dentada", "Gates", "Correia dentada com kit de tensores", 210.0, 8, 2, admin.ID)

	criarOrdemServicoEmAndamento(ctx, c, clientePrincipal.ID, veiculo1.ID, atendente.ID, mecanico.ID)

	imprimirCredenciais(credenciais)
}

func criarUsuarioInterno(ctx context.Context, c *wiring.Container, nome, email, papel string) appusuario.UsuarioOutput {
	out, err := c.CriarUsuarioUC.Executar(ctx, appusuario.CriarUsuarioInput{
		Nome:         nome,
		Email:        email,
		SenhaInicial: senhaPadrao,
		Papel:        papel,
	})
	if err != nil {
		log.Fatalf("erro ao criar usuario %s: %v", email, err)
	}
	return out
}

func criarCliente(ctx context.Context, c *wiring.Container, documento, nome, email, telefone string, criadoPor uint64) appcliente.ClienteOutput {
	out, err := c.CriarClienteUseCase.Executar(ctx, appcliente.CriarClienteInput{
		Documento: documento,
		Tipo:      "PF",
		Nome:      nome,
		Email:     email,
		Telefone:  telefone,
		Senha:     senhaPadrao,
		CriadoPor: criadoPor,
	})
	if err != nil {
		log.Fatalf("erro ao criar cliente %s: %v", email, err)
	}
	return out
}

func criarVeiculo(ctx context.Context, c *wiring.Container, placa, marca, modelo string, km uint32, ano uint16, cor string, criadoPor uint64) appveiculo.VeiculoOutput {
	out, err := c.CadastrarVeiculoUC.Executar(ctx, appveiculo.CadastrarVeiculoInput{
		Placa:              placa,
		Marca:              marca,
		Modelo:             modelo,
		QuilometragemAtual: km,
		Ano:                ano,
		Cor:                cor,
		CriadoPor:          criadoPor,
	})
	if err != nil {
		log.Fatalf("erro ao criar veiculo %s: %v", placa, err)
	}
	return out
}

func criarServico(ctx context.Context, c *wiring.Container, nome, descricao string, preco float64, tempoMinutos int, criadoPor uint64) {
	if _, err := c.CriarServicoUC.Executar(ctx, appservico.CriarServicoInput{
		Nome:                 nome,
		Descricao:            descricao,
		PrecoBase:            preco,
		TempoEstimadoMinutos: tempoMinutos,
		CriadoPor:            criadoPor,
	}); err != nil {
		log.Fatalf("erro ao criar servico %s: %v", nome, err)
	}
}

func criarPeca(ctx context.Context, c *wiring.Container, nome, marca, descricao string, preco float64, quantidade, estoqueMinimo int, criadoPor uint64) {
	if _, err := c.CadastrarPecaUC.Executar(ctx, apppeca.CadastrarPecaInput{
		Nome:                nome,
		Marca:               marca,
		Descricao:           descricao,
		Preco:               preco,
		QuantidadeEmEstoque: quantidade,
		EstoqueMinimo:       estoqueMinimo,
		CriadoPor:           criadoPor,
	}); err != nil {
		log.Fatalf("erro ao criar peca %s: %v", nome, err)
	}
}

// criarOrdemServicoEmAndamento leva a OS até EM_DIAGNOSTICO com diagnóstico
// preenchido, o estado mais avançado alcançável hoje: não existe use case de
// aprovação (EM_DIAGNOSTICO -> AGUARDANDO_APROVACAO -> APROVADA), pré-requisito
// da transição para EM_EXECUCAO.
func criarOrdemServicoEmAndamento(ctx context.Context, c *wiring.Container, clienteID, veiculoID, atendenteID, mecanicoID uint64) {
	os, err := c.AbrirOrdemServicoUC.Executar(ctx, appordemservico.AbrirOrdemServicoInput{
		ClienteID:            clienteID,
		VeiculoID:            veiculoID,
		QuilometragemEntrada: 45000,
		Observacoes:          "Cliente relata barulho na suspensão dianteira",
		UsuarioID:            atendenteID,
	})
	if err != nil {
		log.Fatalf("erro ao abrir ordem de servico: %v", err)
	}

	if _, err := c.IniciarDiagnosticoUC.Executar(ctx, appordemservico.IniciarDiagnosticoInput{
		OrdemServicoID: os.ID,
		UsuarioID:      mecanicoID,
	}); err != nil {
		log.Fatalf("erro ao iniciar diagnostico: %v", err)
	}

	if _, err := c.InformarDiagnosticoUC.Executar(ctx, appordemservico.InformarDiagnosticoInput{
		OrdemServicoID: os.ID,
		Diagnostico:    "Amortecedores dianteiros desgastados, necessária substituição",
	}); err != nil {
		log.Fatalf("erro ao informar diagnostico: %v", err)
	}
}

func imprimirCredenciais(credenciais []credencial) {
	fmt.Println()
	fmt.Println("========================================================")
	fmt.Println(" DADOS DE LOGIN PARA AVALIAÇÃO")
	fmt.Println("========================================================")
	for _, cr := range credenciais {
		fmt.Printf(" papel: %-10s nome: %-20s email: %-25s senha: %s\n", cr.papel, cr.nome, cr.email, cr.senha)
	}
	fmt.Println("========================================================")
	fmt.Println()
}
