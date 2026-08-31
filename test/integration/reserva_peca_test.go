//go:build integration

package integration_test

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"testing"
	"time"

	apporcamento "github.com/muriloperosa/soat-architecture/internal/application/orcamento"
	appordemservico "github.com/muriloperosa/soat-architecture/internal/application/ordemservico"
	apppeca "github.com/muriloperosa/soat-architecture/internal/application/peca"
	domainpeca "github.com/muriloperosa/soat-architecture/internal/domain/peca"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httppeca "github.com/muriloperosa/soat-architecture/internal/infrastructure/http/peca"
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
	require.NoError(t, err)
	return out.ID
}

func clienteIDDaOS(t *testing.T, ordemServicoID uint64) uint64 {
	t.Helper()
	var clienteID uint64
	require.NoError(t, testDB.Table("ordens_servico").
		Select("cliente_id").
		Where("id = ?", ordemServicoID).
		Scan(&clienteID).Error)
	require.NotZero(t, clienteID)
	return clienteID
}

func prepararOrcamentoComPecaAguardando(
	t *testing.T,
	ordemServicoID, usuarioID, pecaID uint64,
	quantidade int,
) {
	t.Helper()
	ctx := context.Background()

	_, err := testContainer.IniciarDiagnosticoUC.Executar(ctx, appordemservico.IniciarDiagnosticoInput{
		OrdemServicoID: ordemServicoID,
		UsuarioID:      usuarioID,
	})
	require.NoError(t, err)

	_, err = testContainer.InformarDiagnosticoUC.Executar(ctx, appordemservico.InformarDiagnosticoInput{
		OrdemServicoID: ordemServicoID,
		Diagnostico:    "Peça precisa ser substituída",
	})
	require.NoError(t, err)

	_, err = testContainer.GerarOrcamentoUC.Executar(ctx, apporcamento.GerarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		Observacoes:    "Orçamento para aprovação",
		UsuarioID:      usuarioID,
	})
	require.NoError(t, err)

	_, err = testContainer.AdicionarPecaOrcamentoUC.Executar(ctx, apporcamento.AdicionarPecaOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		PecaID:         pecaID,
		Quantidade:     quantidade,
	})
	require.NoError(t, err)

	_, err = testContainer.EnviarOrcamentoParaAprovacaoUC.Executar(ctx, apporcamento.EnviarOrcamentoParaAprovacaoInput{
		OrdemServicoID: ordemServicoID,
		UsuarioID:      usuarioID,
	})
	require.NoError(t, err)
}

func TestAprovarOrcamento_CriaReservaAutomaticamenteComQuantidadeDoOrcamento(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 2)
	ordemServicoID := seedOrdemServico(t, admin.ID)
	clienteID := clienteIDDaOS(t, ordemServicoID)
	prepararOrcamentoComPecaAguardando(t, ordemServicoID, admin.ID, pecaID, 3)

	out, err := testContainer.AprovarOrcamentoUC.Executar(context.Background(), apporcamento.AprovarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ClienteID:      clienteID,
	})

	require.NoError(t, err)
	require.Equal(t, "APROVADA", out.Status)

	reserva, err := testContainer.ReservaPecaRepo.BuscarPorOrdemEPeca(context.Background(), ordemServicoID, pecaID)
	require.NoError(t, err)
	require.Equal(t, 3, reserva.Quantidade())
}

func TestAprovarOrcamento_EstoqueInsuficienteFazRollbackCompleto(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	pecaID := seedPecaComEstoque(t, admin.ID, 5, 2)
	ordemServicoID := seedOrdemServico(t, admin.ID)
	clienteID := clienteIDDaOS(t, ordemServicoID)
	prepararOrcamentoComPecaAguardando(t, ordemServicoID, admin.ID, pecaID, 4)

	_, err := testContainer.AprovarOrcamentoUC.Executar(context.Background(), apporcamento.AprovarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ClienteID:      clienteID,
	})

	require.ErrorIs(t, err, domainpeca.ErrQuantidadeIndisponivelParaReserva)

	var status string
	require.NoError(t, testDB.Table("ordens_servico").Select("status").Where("id = ?", ordemServicoID).Scan(&status).Error)
	require.Equal(t, "AGUARDANDO_APROVACAO", status, "falha na reserva não pode aprovar a OS")

	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Zero(t, reservada, "falha na aprovação deve fazer rollback das reservas")
}

func TestAprovarOrcamento_ConcorrenciaNaoUltrapassaEstoqueDisponivel(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	pecaID := seedPecaComEstoque(t, admin.ID, 5, 0)

	// Cria uma primeira OS para obter cliente e veículo válidos e reaproveita
	// ambos nas demais OSs. O que varia é o orçamento/OS concorrente.
	primeiraOS := seedOrdemServico(t, admin.ID)
	var base struct {
		ClienteID uint64 `gorm:"column:cliente_id"`
		VeiculoID uint64 `gorm:"column:veiculo_id"`
	}
	require.NoError(t, testDB.Table("ordens_servico").
		Select("cliente_id, veiculo_id").
		Where("id = ?", primeiraOS).
		Scan(&base).Error)

	const tentativas = 10
	ordens := make([]uint64, 0, tentativas)
	ordens = append(ordens, primeiraOS)

	for i := 1; i < tentativas; i++ {
		numero := fmt.Sprintf("OSC%06d%02d", time.Now().UnixNano()%1_000_000, i)
		result := testDB.Exec(
			"INSERT INTO ordens_servico (numero, cliente_id, veiculo_id, quilometragem_entrada, status, criado_por) VALUES (?, ?, ?, 0, 'RECEBIDA', ?)",
			numero, base.ClienteID, base.VeiculoID, admin.ID,
		)
		require.NoError(t, result.Error)

		var id uint64
		require.NoError(t, testDB.Raw("SELECT id FROM ordens_servico WHERE numero = ?", numero).Scan(&id).Error)
		ordens = append(ordens, id)
	}

	for _, ordemServicoID := range ordens {
		prepararOrcamentoComPecaAguardando(t, ordemServicoID, admin.ID, pecaID, 1)
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	sucessos := 0

	for _, ordemServicoID := range ordens {
		ordemServicoID := ordemServicoID
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := testContainer.AprovarOrcamentoUC.Executar(context.Background(), apporcamento.AprovarOrcamentoInput{
				OrdemServicoID: ordemServicoID,
				ClienteID:      base.ClienteID,
			})
			mu.Lock()
			defer mu.Unlock()
			if err == nil {
				sucessos++
			}
		}()
	}
	wg.Wait()

	require.Equal(t, 5, sucessos, "somente cinco aprovações podem reservar uma unidade cada")
	reservada, err := testContainer.ReservaPecaRepo.SomarQuantidadeReservada(context.Background(), pecaID)
	require.NoError(t, err)
	require.Equal(t, 5, reservada)
}

func TestAprovarOrcamento_GetPecaReflemeteSaldoDisponivelDescontandoReserva(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")
	pecaID := seedPecaComEstoque(t, admin.ID, 10, 2)
	ordemServicoID := seedOrdemServico(t, admin.ID)
	clienteID := clienteIDDaOS(t, ordemServicoID)
	prepararOrcamentoComPecaAguardando(t, ordemServicoID, admin.ID, pecaID, 3)

	_, err := testContainer.AprovarOrcamentoUC.Executar(context.Background(), apporcamento.AprovarOrcamentoInput{
		OrdemServicoID: ordemServicoID,
		ClienteID:      clienteID,
	})
	require.NoError(t, err)

	var resposta httppeca.PecaResponse
	rec := doRequest(t, http.MethodGet, "/v1/pecas/"+strconv.FormatUint(pecaID, 10), login.AccessToken, nil, &resposta)
	require.Equal(t, http.StatusOK, rec.Code)

	require.Equal(t, 10, resposta.QuantidadeEmEstoque, "estoque físico não deve mudar por causa da reserva")
	require.Equal(t, 3, resposta.QuantidadeReservada)
	require.Equal(t, 7, resposta.QuantidadeDisponivel, "saldo disponível deve descontar a reserva da OS aprovada")
}

func TestReservaPeca_NaoPossuiEndpointManual(t *testing.T) {
	resetDB(t)
	admin := seedUsuario(t, "Admin Oficina", "admin@oficina.com", "senha123", shared.PapelAdmin)
	login := doLogin(t, "admin@oficina.com", "senha123")
	ordemServicoID := seedOrdemServico(t, admin.ID)

	rec := doRequest(
		t,
		http.MethodPost,
		"/v1/ordens-servico/"+strconv.FormatUint(ordemServicoID, 10)+"/reservas-pecas",
		login.AccessToken,
		map[string]any{"peca_id": 1, "quantidade": 1},
		nil,
	)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
