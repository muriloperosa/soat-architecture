package reservapeca

import "context"

// Repository persiste e consulta reservas de peça. Segue o fluxo
// transacional descrito em docs/database/DATABASE.md: quem chama isso
// (futuro OrdemServico.adicionarPeca/removerPeca) é responsável por
// bloquear a peça (SELECT ... FOR UPDATE em peca.Repository) antes de
// checar SomarQuantidadeReservada e decidir Salvar ou Atualizar.
type Repository interface {
	Salvar(ctx context.Context, reserva *ReservaPeca) error
	Atualizar(ctx context.Context, reserva *ReservaPeca) error
	BuscarPorOrdemEPeca(ctx context.Context, ordemServicoID, pecaID uint64) (*ReservaPeca, error)
	BuscarPorOrdemServico(ctx context.Context, ordemServicoID uint64) ([]*ReservaPeca, error)
	SomarQuantidadeReservada(ctx context.Context, pecaID uint64) (int, error)
	Remover(ctx context.Context, ordemServicoID, pecaID uint64) error
}
