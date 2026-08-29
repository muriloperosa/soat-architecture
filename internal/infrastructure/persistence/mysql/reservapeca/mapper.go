package reservapeca

import domain "github.com/muriloperosa/soat-architecture/internal/domain/reservapeca"

func toModel(reserva *domain.ReservaPeca) *ReservaPecaModel {
	return &ReservaPecaModel{
		ID:             reserva.ID(),
		OrdemServicoID: reserva.OrdemServicoID(),
		PecaID:         reserva.PecaID(),
		Quantidade:     reserva.Quantidade(),
		CriadaEm:       reserva.CriadaEm(),
		AtualizadaEm:   reserva.AtualizadaEm(),
	}
}

func toDomain(model ReservaPecaModel) *domain.ReservaPeca {
	return domain.RestaurarReservaPeca(
		model.ID,
		model.OrdemServicoID,
		model.PecaID,
		model.Quantidade,
		model.CriadaEm,
		model.AtualizadaEm,
	)
}
