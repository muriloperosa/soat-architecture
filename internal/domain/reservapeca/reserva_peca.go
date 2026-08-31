package reservapeca

import "time"

// ReservaPeca representa a quantidade de uma Peca comprometida com uma
// Ordem de Serviço. Não é aggregate root: é uma entidade interna do
// agregado OrdemServico. Ela é criada automaticamente quando o cliente
// aprova o orçamento. Uma OS possui no máximo uma reserva por peça
// (UNIQUE ordem_servico_id+peca_id na migration).
// Disponibilidade de uma peça = estoque físico - soma das reservas.
type ReservaPeca struct {
	id             uint64
	ordemServicoID uint64
	pecaID         uint64
	quantidade     int
	criadaEm       time.Time
	atualizadaEm   time.Time
}

func NewReservaPeca(ordemServicoID, pecaID uint64, quantidade int) (*ReservaPeca, error) {
	if ordemServicoID == 0 {
		return nil, ErrOrdemServicoObrigatoria
	}

	if pecaID == 0 {
		return nil, ErrPecaObrigatoria
	}

	if quantidade <= 0 {
		return nil, ErrQuantidadeInvalida
	}

	agora := time.Now()

	return &ReservaPeca{
		ordemServicoID: ordemServicoID,
		pecaID:         pecaID,
		quantidade:     quantidade,
		criadaEm:       agora,
		atualizadaEm:   agora,
	}, nil
}

// RestaurarReservaPeca reidrata uma ReservaPeca a partir de dados já
// persistidos; não reaplica validação de negócio. Usado só pelo mapper de
// persistência (internal/infrastructure/persistence/mysql/reservapeca).
func RestaurarReservaPeca(id, ordemServicoID, pecaID uint64, quantidade int, criadaEm, atualizadaEm time.Time) *ReservaPeca {
	return &ReservaPeca{
		id:             id,
		ordemServicoID: ordemServicoID,
		pecaID:         pecaID,
		quantidade:     quantidade,
		criadaEm:       criadaEm,
		atualizadaEm:   atualizadaEm,
	}
}

// A quantidade da reserva nasce do orçamento aprovado e não é alterada
// diretamente. Qualquer mudança de quantidade deve acontecer no orçamento,
// invalidar a aprovação anterior e gerar novas reservas após nova aprovação.

func (r *ReservaPeca) AtribuirID(id uint64) {
	r.id = id
}

func (r *ReservaPeca) ID() uint64              { return r.id }
func (r *ReservaPeca) OrdemServicoID() uint64  { return r.ordemServicoID }
func (r *ReservaPeca) PecaID() uint64          { return r.pecaID }
func (r *ReservaPeca) Quantidade() int         { return r.quantidade }
func (r *ReservaPeca) CriadaEm() time.Time     { return r.criadaEm }
func (r *ReservaPeca) AtualizadaEm() time.Time { return r.atualizadaEm }
