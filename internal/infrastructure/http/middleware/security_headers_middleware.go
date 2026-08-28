package middleware

import "github.com/gin-gonic/gin"

// SecurityHeadersMiddleware seta headers de segurança em toda resposta,
// independente de rota: X-Content-Type-Options impede MIME-sniffing em
// navegadores antigos, Cross-Origin-Resource-Policy restringe quem pode
// embutir a resposta como recurso cross-origin.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("Cross-Origin-Resource-Policy", "same-origin")
		c.Next()
	}
}
