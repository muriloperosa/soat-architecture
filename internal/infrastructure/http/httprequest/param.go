// Package httprequest reúne helpers de leitura de request HTTP (path
// params, etc.) reusados entre handlers de domínios diferentes.
package httprequest

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

// ParseUintParam lê e valida o path param name como uint64. Em caso de
// erro, já responde 400 (RespondValidationError) e devolve ok=false; o
// handler chamador só precisa checar ok e retornar.
func ParseUintParam(c *gin.Context, name string) (uint64, bool) {
	valor, err := strconv.ParseUint(c.Param(name), 10, 64)
	if err != nil {
		httperror.RespondValidationError(c, name+" inválido.")
		return 0, false
	}
	return valor, true
}
