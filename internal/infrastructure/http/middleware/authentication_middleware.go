package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

// ClaimsContextKey é a chave usada pra guardar *domainauth.AppClaims no gin.Context.
const ClaimsContextKey = "auth.claims"

// AuthenticationMiddleware valida assinatura e expiração do JWT recebido no
// header Authorization (Bearer) e injeta os claims tipados no contexto.
func AuthenticationMiddleware(jwtAuth domainauth.JWTProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		tokenBruto, ok := strings.CutPrefix(header, "Bearer ")
		if !ok || tokenBruto == "" {
			httperror.RespondUnauthorizedError(c, "Requisição não autorizada.")
			c.Abort()
			return
		}

		claims, err := jwtAuth.ValidarAccessToken(tokenBruto)
		if err != nil {
			httperror.RespondUnauthorizedError(c, "Requisição não autorizada.")
			c.Abort()
			return
		}

		c.Set(ClaimsContextKey, claims)
		c.Next()
	}
}
