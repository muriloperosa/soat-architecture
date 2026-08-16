package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/domain/shared"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

// AuthorizationMiddleware exige que o AppClaims injetado por
// AuthenticationMiddleware tenha o TipoUsuario esperado pra essa rota, e,
// se papeisPermitidos não estiver vazio, que o Papel esteja entre eles.
func AuthorizationMiddleware(tipoEsperado domainauth.TipoUsuario, papeisPermitidos ...shared.PapelUsuario) gin.HandlerFunc {
	return func(c *gin.Context) {
		valor, existe := c.Get(ClaimsContextKey)
		claims, ok := valor.(*domainauth.AppClaims)
		if !existe || !ok || claims.Tipo != tipoEsperado {
			httperror.RespondForbiddenError(c, "Acesso não permitido para este tipo de usuário.")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		if len(papeisPermitidos) > 0 && !papelPermitido(claims.Papel, papeisPermitidos) {
			httperror.RespondForbiddenError(c, "Acesso não permitido para este papel de usuário.")
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

// papelPermitido indica se papel está entre os permitidos.
func papelPermitido(papel shared.PapelUsuario, permitidos []shared.PapelUsuario) bool {
	for _, p := range permitidos {
		if papel == p {
			return true
		}
	}
	return false
}
