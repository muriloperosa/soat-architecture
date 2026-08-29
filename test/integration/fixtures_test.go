//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	appcliente "github.com/muriloperosa/soat-architecture/internal/application/cliente"
	appusuario "github.com/muriloperosa/soat-architecture/internal/application/usuario"
	domaincliente "github.com/muriloperosa/soat-architecture/internal/domain/cliente"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
)

func resetDB(t *testing.T) {
	t.Helper()
	require := func(err error) {
		if err != nil {
			t.Fatalf("erro ao limpar banco de teste: %v", err)
		}
	}
	require(testDB.Exec("DELETE FROM refresh_tokens").Error)
	require(testDB.Exec("DELETE FROM reservas_pecas").Error)
	require(testDB.Exec("DELETE FROM orcamentos_pecas").Error)
	require(testDB.Exec("DELETE FROM orcamentos_servicos").Error)
	require(testDB.Exec("DELETE FROM orcamentos").Error)
	require(testDB.Exec("DELETE FROM historicos_status").Error)
	require(testDB.Exec("DELETE FROM ordens_servico").Error)
	require(testDB.Exec("DELETE FROM clientes").Error)
	require(testDB.Exec("DELETE FROM pecas").Error)
	require(testDB.Exec("DELETE FROM veiculos").Error)
	require(testDB.Exec("DELETE FROM usuarios").Error)
}

func seedUsuario(t *testing.T, nome, email, senha string, papel shared.PapelUsuario) appusuario.UsuarioOutput {
	t.Helper()
	out, err := testContainer.CriarUsuarioUC.Executar(context.Background(), appusuario.CriarUsuarioInput{
		Nome:         nome,
		Email:        email,
		SenhaInicial: senha,
		Papel:        string(papel),
	})
	if err != nil {
		t.Fatalf("erro ao semear usuário de teste: %v", err)
	}
	return out
}

// seedOrdemServico cria os pré-requisitos de uma Ordem de Serviço (cliente
// via use case real; veículo e a própria OS via SQL cru, porque os
// domínios veiculo/ordemservico ainda não existem) só pra satisfazer a FK
// reservas_pecas.ordem_servico_id -> ordens_servico.id nos testes de
// reserva. Retorna o ID da OS semeada.
func seedOrdemServico(t *testing.T, criadoPor uint64) uint64 {
	t.Helper()

	sufixo := time.Now().UnixNano() % 1_000_000

	clienteOut, err := testContainer.CriarClienteUseCase.Executar(context.Background(), appcliente.CriarClienteInput{
		Documento: "52998224725",
		Tipo:      string(domaincliente.TipoPessoaFisica),
		Nome:      "Cliente Teste Reserva",
		Email:     fmt.Sprintf("cliente-reserva-%d@teste.com", sufixo),
		Telefone:  "44999991234",
		Senha:     "senha123",
		CriadoPor: criadoPor,
	})
	if err != nil {
		t.Fatalf("erro ao semear cliente pra OS de teste: %v", err)
	}

	placa := fmt.Sprintf("V%06d", sufixo)
	if err := testDB.Exec(
		"INSERT INTO veiculos (placa, marca, modelo, quilometragem_atual, ano, cor, criado_por) VALUES (?, 'Fiat', 'Uno', 0, 2020, 'Branco', ?)",
		placa, criadoPor,
	).Error; err != nil {
		t.Fatalf("erro ao semear veiculo pra OS de teste: %v", err)
	}

	var veiculoID uint64
	if err := testDB.Raw("SELECT id FROM veiculos WHERE placa = ?", placa).Scan(&veiculoID).Error; err != nil {
		t.Fatalf("erro ao buscar veiculo semeado: %v", err)
	}

	numero := fmt.Sprintf("OS%06d", sufixo)
	if err := testDB.Exec(
		"INSERT INTO ordens_servico (numero, cliente_id, veiculo_id, quilometragem_entrada, status, criado_por) VALUES (?, ?, ?, 0, 'RECEBIDA', ?)",
		numero, clienteOut.ID, veiculoID, criadoPor,
	).Error; err != nil {
		t.Fatalf("erro ao semear ordem de serviço de teste: %v", err)
	}

	var ordemServicoID uint64
	if err := testDB.Raw("SELECT id FROM ordens_servico WHERE numero = ?", numero).Scan(&ordemServicoID).Error; err != nil {
		t.Fatalf("erro ao buscar ordem de serviço semeada: %v", err)
	}

	return ordemServicoID
}
