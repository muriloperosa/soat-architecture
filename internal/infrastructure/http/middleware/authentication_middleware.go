package middleware

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	domainauth "github.com/muriloperosa/soat-architecture/internal/domain/auth"
	"github.com/muriloperosa/soat-architecture/internal/infrastructure/http/httperror"
)

// ClaimsContextKey é a chave usada pra guardar *domainauth.AppClaims no gin.Context.
const ClaimsContextKey = "auth.claims"

// AuthenticationMiddleware valida assinatura e expiração do JWT recebido no
// header Authorization (Bearer), rejeita se o access token foi revogado em
// par (logout ou rotação do refresh token correspondente), rejeita se o
// usuário/cliente foi inativado desde a emissão do token (checado a cada
// request, não só no login) e injeta os claims tipados no contexto.
//
// A checagem de "ativo" consulta usuarios ou clientes de acordo com
// claims.Tipo (extraído do próprio token, não da rota) — por isso recebe os
// dois repositórios, e não só o esperado pra essa rota: se dependesse do
// tipo esperado da rota, um token do tipo errado seria checado contra a
// tabela errada antes de AuthorizationMiddleware barrar o Tipo, retornando
// 401 por coincidência (ou pior, achando um registro não relacionado com o
// mesmo ID) em vez do 403 correto.
func AuthenticationMiddleware(jwtAuth domainauth.JWTProvider, refreshTokens domainauth.RefreshTokenRepository, usuarios, clientes domainauth.UsuarioStatusRepository) gin.HandlerFunc {
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

		revogado, err := refreshTokens.AccessTokenRevogado(c.Request.Context(), claims.Jti)
		if err != nil || revogado {
			httperror.RespondUnauthorizedError(c, "Requisição não autorizada.")
			c.Abort()
			return
		}

		subjectID, err := strconv.ParseUint(claims.Subject, 10, 64)
		if err != nil {
			httperror.RespondUnauthorizedError(c, "Requisição não autorizada.")
			c.Abort()
			return
		}

		status := usuarios
		if claims.Tipo == domainauth.TipoCliente {
			status = clientes
		}

		ativo, err := status.EstaAtivo(c.Request.Context(), subjectID)
		if err != nil || !ativo {
			httperror.RespondUnauthorizedError(c, "Requisição não autorizada.")
			c.Abort()
			return
		}

		c.Set(ClaimsContextKey, claims)
		c.Next()
	}
}
