package auth

// JWTProvider emite e valida o access token JWT, e emite o valor bruto do
// refresh token. Implementado em internal/infrastructure/auth (AuthenticatorJWT).
type JWTProvider interface {
	GerarAccessToken(subject string, tipo TipoUsuario, papel PapelUsuario) (token string, jti string, err error)
	ValidarAccessToken(tokenBruto string) (*AppClaims, error)
	GerarRefreshToken() (string, error)
}
