package relatorio

import (
	"database/sql"
	"time"

	domain "github.com/muriloperosa/soat-architecture/internal/domain/relatorio"
)

// transicaoStatusResultado recebe o resultado escalar da query de agregação.
// Os campos ficam NULL quando não há nenhuma OS que completou a transição
// no período (AVG/MIN/MAX de zero linhas).
type transicaoStatusResultado struct {
	Total          int64
	MediaSegundos  sql.NullFloat64
	MinimaSegundos sql.NullFloat64
	MaximaSegundos sql.NullFloat64
}

func toDomain(resultado transicaoStatusResultado) domain.TransicaoStatusResultado {
	return domain.TransicaoStatusResultado{
		TotalOrdens:   int(resultado.Total),
		DuracaoMedia:  secondsToDuration(resultado.MediaSegundos),
		DuracaoMinima: secondsToDuration(resultado.MinimaSegundos),
		DuracaoMaxima: secondsToDuration(resultado.MaximaSegundos),
	}
}

func secondsToDuration(seconds sql.NullFloat64) time.Duration {
	if !seconds.Valid {
		return 0
	}
	return time.Duration(seconds.Float64) * time.Second
}
