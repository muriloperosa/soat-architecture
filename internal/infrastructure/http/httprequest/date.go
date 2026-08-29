package httprequest

import (
	"time"

	"github.com/gin-gonic/gin"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

const formatoDataQueryParam = "2006-01-02"

// ParseDateQueryParam lê e valida o query param name como data no formato
// YYYY-MM-DD, sempre em UTC. Em caso de erro (ausente ou formato inválido),
// já responde 400 (RespondValidationError) e devolve ok=false; o handler
// chamador só precisa checar ok e retornar.
func ParseDateQueryParam(c *gin.Context, name string) (time.Time, bool) {
	valor := c.Query(name)
	data, err := time.ParseInLocation(formatoDataQueryParam, valor, time.UTC)
	if err != nil {
		httperror.RespondValidationError(c, name+" inválido, formato esperado YYYY-MM-DD.")
		return time.Time{}, false
	}
	return data, true
}
