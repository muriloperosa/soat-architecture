package relatorio

import (
	"context"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
}

var _ domain.RelatorioTransicaoStatusRepository = (*Repository)(nil)

func NewRelatorioRepository(db *gorm.DB) domain.RelatorioTransicaoStatusRepository {
	return &Repository{db: db}
}

// consultaTransicaoStatus: fim é a data (início do dia) informada pelo
// usuário; o dia inteiro do final_date entra na busca via limite exclusivo
// do dia seguinte (t_to < fim + 1 dia).
const consultaTransicaoStatus = `
WITH primeiro_from AS (
	SELECT ordem_servico_id, MIN(alterado_em) AS t_from
	FROM historicos_status
	WHERE status = ?
	GROUP BY ordem_servico_id
),
primeiro_to_apos_from AS (
	SELECT pf.ordem_servico_id, pf.t_from, MIN(hs.alterado_em) AS t_to
	FROM primeiro_from pf
	JOIN historicos_status hs
		ON hs.ordem_servico_id = pf.ordem_servico_id
	   AND hs.status = ?
	   AND hs.alterado_em > pf.t_from
	GROUP BY pf.ordem_servico_id, pf.t_from
)
SELECT
	COUNT(*) AS total,
	AVG(TIMESTAMPDIFF(SECOND, t_from, t_to)) AS media_segundos,
	MIN(TIMESTAMPDIFF(SECOND, t_from, t_to)) AS minima_segundos,
	MAX(TIMESTAMPDIFF(SECOND, t_from, t_to)) AS maxima_segundos
FROM primeiro_to_apos_from
WHERE t_to >= ? AND t_to < DATE_ADD(?, INTERVAL 1 DAY)
`

func (r *Repository) CalcularTransicaoStatus(ctx context.Context, params domain.CalcularTransicaoStatusParams) (domain.TransicaoStatusResultado, error) {
	var resultado transicaoStatusResultado

	err := r.db.WithContext(ctx).Raw(
		consultaTransicaoStatus,
		params.FromStatus.String(),
		params.ToStatus.String(),
		params.Periodo.Inicio(),
		params.Periodo.Fim(),
	).Scan(&resultado).Error
	if err != nil {
		return domain.TransicaoStatusResultado{}, err
	}

	return toDomain(resultado), nil
}
