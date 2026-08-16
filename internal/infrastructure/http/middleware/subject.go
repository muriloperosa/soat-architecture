package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

// SubjectID lê o ID do usuário logado a partir dos claims injetados por
// AuthenticationMiddleware. Usado por endpoints self-service (ex.
// GET /me, PUT /me/senha) que nunca recebem :id, só operam sobre o próprio
// usuário do token.
func SubjectID(c *gin.Context) (uint64, bool) {
	valor, existe := c.Get(ClaimsContextKey)
	claims, ok := valor.(*domainauth.AppClaims)
	if !existe || !ok {
		httperror.RespondUnauthorizedError(c, "Requisição não autorizada.")
		return 0, false
	}
	id, err := strconv.ParseUint(claims.Subject, 10, 64)
	if err != nil {
		httperror.RespondUnauthorizedError(c, "Requisição não autorizada.")
		return 0, false
	}
	return id, true
}
