package reservapeca

import "context"

// Repository persiste e consulta reservas geradas automaticamente a partir
// de orçamentos aprovados. A quantidade da reserva não possui operação de
// atualização direta: mudanças devem ocorrer no orçamento e gerar uma nova
// aprovação/reserva.
type Repository interface {
	Salvar(ctx context.Context, reserva *ReservaPeca) error
	BuscarPorOrdemEPeca(ctx context.Context, ordemServicoID, pecaID uint64) (*ReservaPeca, error)
	BuscarPorOrdemServico(ctx context.Context, ordemServicoID uint64) ([]*ReservaPeca, error)
	SomarQuantidadeReservada(ctx context.Context, pecaID uint64) (int, error)
	Remover(ctx context.Context, ordemServicoID, pecaID uint64) error
}
