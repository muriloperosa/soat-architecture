package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

// RecoveryMiddleware substitui o gin.Recovery() padrão, que devolve texto
// plano "Internal Server Error" em panic. Aqui a resposta segue o mesmo
// contrato JSON do resto da API (httperror.RespondError), sem vazar o
// valor do panic (stack trace continua só no log via gin.Logger()).
func RecoveryMiddleware() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered any) {
		err, ok := recovered.(error)
		if !ok {
			err = fmt.Errorf("%v", recovered)
		}

		httperror.RespondInternalError(c, "erro interno", err)
		c.Abort()
	})
}
