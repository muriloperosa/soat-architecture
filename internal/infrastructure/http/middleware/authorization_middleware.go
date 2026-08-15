package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	httphandler "github.com/muriloperosa/soat-architecture/internal/infrastructure/http"
)

const msgAcessoNaoPermitido = "acesso não permitido para este tipo de usuário"

// AuthorizationMiddleware exige que o AppClaims injetado por
// AuthenticationMiddleware tenha o TipoUsuario esperado pra essa rota.
func AuthorizationMiddleware(tipoEsperado domainauth.TipoUsuario) gin.HandlerFunc {
	return func(c *gin.Context) {
		valor, existe := c.Get(ClaimsContextKey)
		claims, ok := valor.(*domainauth.AppClaims)
		if !existe || !ok || claims.Tipo != tipoEsperado {
			httphandler.RespondError(c, shared.NewForbiddenError(msgAcessoNaoPermitido))
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
