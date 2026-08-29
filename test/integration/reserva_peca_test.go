//go:build integration

package integration_test

import (
	"context"
	"sync"
	"testing"

	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/stretchr/testify/require"
)

func seedPecaComEstoque(t *testing.T, criadoPor uint64, quantidadeEmEstoque, estoqueMinimo int) uint64 {
	t.Helper()

	out, err := testContainer.CadastrarPecaUC.Executar(context.Background(), apppeca.CadastrarPecaInput{
		Nome:                "Pastilha de freio",
		Marca:               "Bosch",
		Descricao:           "Pastilha dianteira",
		Preco:               89.9,
		QuantidadeEmEstoque: quantidadeEmEstoque,
		EstoqueMinimo:       estoqueMinimo,
		CriadoPor:           criadoPor,
	})
	if err != nil {
		t.Fatalf("erro ao semear peça de teste: %v", err)
	}

	return out.ID
}

// TestReservarPeca_ReservaComSucesso_RespeitandoEstoqueMinimo prova, contra
// MySQL real, que reservar uma quantidade dentro do disponível (estoque -
// reservado - mínimo) cria a reserva e ela é somável.
func TestReservarPeca_ReservaComSucesso_RespeitandoEstoqueMinimo(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 2)
	ordemServicoID := seedOrdemServico(t, admin.ID)

	output, err := testContainer.ReservarPecaUC.Executar(context.Background(), apppeca.ReservarPecaInput{
		PecaID:         pecaID,
		OrdemServicoID: ordemServicoID,
		Quantidade:     5,
	})

	require.NoError(t, err)
	require.Equal(t, 5, output.Quantidade)

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 5, reservada)
}

// TestReservarPeca_QuantidadeAcimaDoDisponivel_RetornaErro prova que a
// reserva respeita o estoqueMinimo (não é só estoque - reservado, é
// estoque - reservado - novaQuantidade >= estoqueMinimo).
func TestReservarPeca_QuantidadeAcimaDoDisponivel_RetornaErro(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 5)
	ordemServicoID := seedOrdemServico(t, admin.ID)

	// estoque 10, minimo 5: cabe reservar no máximo 5 (10-0-5=5 >= 5)
	_, err := testContainer.ReservarPecaUC.Executar(context.Background(), apppeca.ReservarPecaInput{
		PecaID:         pecaID,
		OrdemServicoID: ordemServicoID,
		Quantidade:     6,
	})

	require.ErrorIs(t, err, domainpeca.ErrQuantidadeIndisponivelParaReserva)

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Zero(t, reservada, "reserva rejeitada não pode ter deixado nenhuma linha")
}

// TestReservarPeca_Concorrencia_NaoUltrapassaEstoqueDisponivel é o teste
// central desta task: dispara várias reservas concorrentes que, se não
// houvesse lock, somariam mais do que o disponível. Com
// BuscarPorIDComBloqueio (FOR UPDATE) + transação, a soma final não pode
// ultrapassar o estoque.
func TestReservarPeca_Concorrencia_NaoUltrapassaEstoqueDisponivel(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	// estoque 5, minimo 0: só cabem 5 unidades reservadas no total.
	pecaID := seedPecaComEstoque(t, admin.ID, 5, 0)
	ordemServicoID := seedOrdemServico(t, admin.ID)

	const tentativas = 10
	const quantidadePorTentativa = 1

	var wg sync.WaitGroup
	var mu sync.Mutex
	sucessos := 0

	for i := 0; i < tentativas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := testContainer.ReservarPecaUC.Executar(context.Background(), apppeca.ReservarPecaInput{
				PecaID:         pecaID,
				OrdemServicoID: ordemServicoID,
				Quantidade:     quantidadePorTentativa,
			})

			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				sucessos++
			}
		}()
	}
	wg.Wait()

	require.Equal(t, 5, sucessos, "exatamente 5 das 10 reservas concorrentes deveriam ter sucesso")

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 5, reservada, "soma reservada não pode ultrapassar o estoque disponível mesmo sob concorrência")
}

// TestLiberarReservaPeca_LiberacaoParcialEDepoisTotal prova o fluxo de
// liberação: reduz a reserva, e ao chegar a zero, remove a linha (a soma
// reservada volta a zero).
func TestLiberarReservaPeca_LiberacaoParcialEDepoisTotal(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 0)
	ordemServicoID := seedOrdemServico(t, admin.ID)

	_, err := testContainer.ReservarPecaUC.Executar(context.Background(), apppeca.ReservarPecaInput{
		PecaID: pecaID, OrdemServicoID: ordemServicoID, Quantidade: 5,
	})
	require.NoError(t, err)

	parcial, err := testContainer.LiberarReservaPecaUC.Executar(context.Background(), apppeca.LiberarReservaPecaInput{
		PecaID: pecaID, OrdemServicoID: ordemServicoID, Quantidade: 2,
	})
	require.NoError(t, err)
	require.Equal(t, 3, parcial.Quantidade)

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 3, reservada)

	total, err := testContainer.LiberarReservaPecaUC.Executar(context.Background(), apppeca.LiberarReservaPecaInput{
		PecaID: pecaID, OrdemServicoID: ordemServicoID, Quantidade: 3,
	})
	require.NoError(t, err)
	require.Zero(t, total.Quantidade)

	reservada, err = testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Zero(t, reservada, "reserva totalmente liberada não deveria deixar linha em reservas_pecas")
}

// TestLiberarReservaPeca_Concorrencia_NaoPerdeAtualizacoes prova a mesma
// proteção do Reservar, agora pro Liberar: Atualizar grava a quantidade
// absoluta calculada em memória, então duas liberações concorrentes na
// mesma reserva, sem lock, fariam a segunda sobrescrever a primeira (lost
// update) — a reserva pararia num valor intermediário em vez de chegar a
// zero, e todas reportariam sucesso incorretamente. Com
// BuscarPorOrdemEPecaComBloqueio + transação, as 5 liberações concorrentes
// de 1 unidade cada serializam e a reserva termina exatamente em zero.
func TestLiberarReservaPeca_Concorrencia_NaoPerdeAtualizacoes(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 0)
	ordemServicoID := seedOrdemServico(t, admin.ID)

	_, err := testContainer.ReservarPecaUC.Executar(context.Background(), apppeca.ReservarPecaInput{
		PecaID: pecaID, OrdemServicoID: ordemServicoID, Quantidade: 5,
	})
	require.NoError(t, err)

	const tentativas = 5

	var wg sync.WaitGroup
	var mu sync.Mutex
	sucessos := 0

	for i := 0; i < tentativas; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			_, err := testContainer.LiberarReservaPecaUC.Executar(context.Background(), apppeca.LiberarReservaPecaInput{
				PecaID: pecaID, OrdemServicoID: ordemServicoID, Quantidade: 1,
			})

			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				sucessos++
			}
		}()
	}
	wg.Wait()

	require.Equal(t, tentativas, sucessos, "as 5 liberações concorrentes de 1 unidade deveriam ter sucesso")

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Zero(t, reservada, "5 liberações de 1 sobre uma reserva de 5 têm que zerar exatamente, sem lost update")
}
